CREATE TABLE public.permission
(
    id          INT PRIMARY KEY,
    name        VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP
);

CREATE TABLE public.role
(
    id          INT PRIMARY KEY,
    name        VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP WITH TIME ZONE

);

CREATE TABLE public.user
(
    id            uuid PRIMARY KEY         DEFAULT uuid_generate_v4(),
    username      VARCHAR(255) NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    auth_type     smallint     NOT NULL,
    role          int          NOT NULL,
    status        smallint     NOT NULL,
    login_history jsonb,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_user_role_id FOREIGN KEY (role) REFERENCES public.role (id)
);

CREATE INDEX IF NOT EXISTS idx_user_username ON public.user (username);
CREATE INDEX IF NOT EXISTS idx_user_email ON public.user (email);

CREATE TABLE public.role_permission
(
    id            INT PRIMARY KEY,
    role_id       INT NOT NULL,
    permission_id INT NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_role_id FOREIGN KEY (role_id) REFERENCES public.role (id),
    CONSTRAINT fk_permission_id FOREIGN KEY (permission_id) REFERENCES public.permission (id)
);

INSERT INTO public.role (id, name, description)
VALUES (1, 'Admin', 'Admin'),
       (2, 'User', 'User');

INSERT INTO public.permission (id, name, description)
VALUES (1, 'Create User', 'Create User permission'),
       (2, 'Update User', 'Update User permission'),
       (3, 'Delete User', 'Delete User permission'),
       (4, 'View User', 'View User permission'),
       (5, 'Manage User Group', 'Manage User Group permission'),
       (6, 'Manage Permission', 'Manage Permission permission'),
       (7, 'Manage User Group Permission', 'Manage User Group Permission permission'),
       (8, 'Manage Self', 'Manage Self permission');

INSERT INTO public.role_permission (id, role_id, permission_id)
VALUES (1, 1, 1),
       (2, 1, 2),
       (3, 1, 3),
       (4, 1, 4),
       (5, 1, 5),
       (6, 1, 6),
       (7, 1, 7),
       (8, 1, 8),
       (9, 2, 8);


