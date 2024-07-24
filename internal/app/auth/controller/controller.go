package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/raffops/auth/internal/app/auth"
	authModels "github.com/raffops/auth/internal/app/auth/model"
	"github.com/raffops/auth/internal/app/user"
	"github.com/raffops/chat/pkg/errs"
	"github.com/raffops/chat/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func init() {
	sanityCheck()
	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_APPLICATION_KEY"),
			os.Getenv("GOOGLE_APPLICATION_SECRET"),
			"http://localhost:8080/login/google/callback",
		),
		github.New(
			os.Getenv("GITHUB_APPLICATION_KEY"),
			os.Getenv("GITHUB_APPLICATION_SECRET"),
			"http://localhost:8080/login/github/callback",
		),
	)

	key := os.Getenv("SESSION_SECRET") // Replace with your SESSION_SECRET or similar
	maxAge := 86400 * 30               // 30 days
	isProd := false                    // SetEncryptedHashmap to true when serving over https

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = isProd

	gothic.Store = store
}

func sanityCheck() {
	envVariables := []string{
		"GOOGLE_APPLICATION_KEY",
		"GOOGLE_APPLICATION_SECRET",
		"GITHUB_APPLICATION_KEY",
		"GITHUB_APPLICATION_SECRET",
		"SESSION_SECRET",
	}

	for _, envVariable := range envVariables {
		if _, ok := os.LookupEnv(envVariable); !ok {
			logger.Fatal("Environment variable not set", zap.String("variable", envVariable))
		}
	}
}

type controller struct {
	userRepo    user.ReaderRepository
	authService auth.Service
}

func (c *controller) SignUp(w http.ResponseWriter, r *http.Request) {
	session, err := gothic.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w,
			errs.NewError(errs.ErrInternal, err).Error(),
			http.StatusInternalServerError,
		)
		return
	}
	ctx := context.Background()

	username := session.Values["username"].(string)
	email := session.Values["email"].(string)
	authType := session.Values["authType"].(string)
	role := session.Values["role"].(string)
	token, errSignup := c.authService.SignUp(ctx, username, email, authType, role)

	if errSignup != nil && errors.Is(errSignup.SvcError(), errs.ErrInternal) {
		http.Error(w, errs.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	if errSignup != nil {
		http.Error(w, errSignup.Error(), errs.GetHttpStatusCode(errSignup))
		return
	}

	response := map[string]any{
		"token": token,
	}

	responseString, _ := json.Marshal(response)
	_, err = w.Write(responseString)
	if err != nil {
		http.Error(w,
			errs.NewError(errs.ErrInternal, err).Error(),
			http.StatusInternalServerError,
		)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (c *controller) Login(w http.ResponseWriter, r *http.Request) {
	session, err := gothic.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w,
			errs.NewError(errs.ErrInternal, err).Error(),
			http.StatusInternalServerError,
		)
		return
	}

	session.Values["username"] = r.URL.Query().Get("username")
	err = session.Save(r, w)
	if err != nil {
		http.Error(w,
			errs.NewError(errs.ErrInternal, err).Error(),
			http.StatusInternalServerError,
		)
		return
	}
	gothic.BeginAuthHandler(w, r)
}

func (c *controller) Callback(w http.ResponseWriter, r *http.Request) {
	session, err := gothic.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w,
			errs.NewError(errs.ErrInternal, err).Error(),
			http.StatusInternalServerError,
		)
		return
	}
	u, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w,
			errs.NewError(errs.ErrInternal, err).Error(),
			http.StatusInternalServerError,
		)
		return
	}
	if u.Email == "" {
		http.Error(w,
			errs.NewError(errs.ErrInternal,
				fmt.Errorf("email not found in %s response", u.Provider),
			).Error(),
			http.StatusInternalServerError,
		)
		return
	}

	ctx := context.Background()
	username := session.Values["username"]
	authType := mux.Vars(r)["provider"]
	getUser, errGetUser := c.userRepo.GetUser(ctx, "username", username)
	if errGetUser != nil && errors.Is(errGetUser.SvcError(), errs.ErrNotFound) {
		redirectToSignUp(w, r, session, u.Email, authType)
		return
	}
	if errGetUser != nil && errors.Is(errGetUser.SvcError(), errs.ErrInternal) {
		http.Error(w, errs.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	token, errLogin := c.authService.Login(ctx, getUser.Username, u.Email)
	if errLogin != nil && errors.Is(errLogin.SvcError(), errs.ErrInternal) {
		http.Error(w, errs.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	if errLogin != nil {
		http.Error(w, errLogin.Error(), errs.GetHttpStatusCode(errLogin))
		return
	}

	response := map[string]any{
		"token": token,
	}
	responseString, _ := json.Marshal(response)
	_, err = w.Write(responseString)
	if err != nil {
		http.Error(w,
			errs.NewError(errs.ErrInternal, err).Error(),
			http.StatusInternalServerError,
		)
	}
	w.WriteHeader(http.StatusOK)
}

func redirectToSignUp(
	w http.ResponseWriter,
	r *http.Request,
	session *sessions.Session,
	email string,
	authType string,
) {
	session.Values["email"] = email
	session.Values["authType"] = authType
	session.Values["role"] = authModels.MapRole[authModels.RoleUser]
	err := session.Save(r, w)
	if err != nil {
		http.Error(
			w,
			errs.NewError(errs.ErrInternal, nil).Error(),
			http.StatusInternalServerError,
		)
		return
	}
	http.Redirect(w, r, "/signUp", http.StatusFound)
}

func NewController(userRepository user.ReaderRepository, authService auth.Service) auth.Controller {
	return &controller{
		userRepo:    userRepository,
		authService: authService,
	}
}
