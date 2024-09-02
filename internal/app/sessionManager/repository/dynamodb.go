package sessionManager

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"github.com/aws/aws-sdk-go/aws"
//	"github.com/aws/aws-sdk-go/aws/awserr"
//	"github.com/aws/aws-sdk-go/service/dynamodb"
//	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
//	"github.com/raffops/auth/internal/app/sessionManager"
//	"github.com/raffops/chat_commons/pkg/encryptor"
//	"github.com/raffops/chat_commons/pkg/errs"
//	"strconv"
//	"time"
//)
//
//type dynamoRepository struct {
//	conn      *dynamodb.DynamoDB
//	encryptor encryptor.Encryptor
//}
//
//func (d dynamoRepository) StringSet(ctx context.Context, tx interface{}, tableName, key, value string) errs.ChatError {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (d dynamoRepository) StringGet(ctx context.Context, tableName, key string) (string, errs.ChatError) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (d dynamoRepository) Delete(ctx context.Context, tableName, key string) errs.ChatError {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (d dynamoRepository) HashSet(
//	ctx context.Context,
//	tableName, key string,
//	values map[string]interface{},
//) errs.ChatError {
//	values["id"] = key
//	marshaledTableItem, err := dynamodbattribute.MarshalMap(values)
//	if err != nil {
//		return errs.NewError(
//			errs.ErrInternal,
//			fmt.Errorf("error marshaling value: %v", err),
//		)
//	}
//
//	input := &dynamodb.PutItemInput{Item: marshaledTableItem, TableName: aws.String(tableName)}
//
//	_, err = d.conn.PutItemWithContext(ctx, input)
//	if err != nil {
//		return errs.NewError(
//			errs.ErrInternal,
//			fmt.Errorf("error creating session: %v", err),
//		)
//	}
//
//	return nil
//}
//
//func (d dynamoRepository) HashSetEncrypted(ctx context.Context, tableName, key, secret string, values map[string]interface{}) errs.ChatError {
//	valueByte, err := json.Marshal(values)
//	if err != nil {
//		return errs.NewError(
//			errs.ErrInternal,
//			fmt.Errorf("error marshaling value: %v", err),
//		)
//	}
//
//	encryptedValue, err := d.encryptor.Encrypt(string(valueByte), secret)
//	if err != nil {
//		return errs.NewError(
//			errs.ErrInternal,
//			fmt.Errorf("error encrypting value: %v", err),
//		)
//	}
//
//	tableItem := map[string]interface{}{
//		"encrypted_value": encryptedValue,
//	}
//
//	return d.HashSet(ctx, tableName, key, tableItem)
//}
//
//func (d dynamoRepository) HashGet(ctx context.Context, tableName, key string, columns ...string) (map[string]interface{}, errs.ChatError) {
//	var attributesToGet []*string
//	for _, column := range columns {
//		attributesToGet = append(attributesToGet, aws.String(column))
//	}
//	input := &dynamodb.GetItemInput{
//		Key: map[string]*dynamodb.AttributeValue{
//			"id": {
//				S: aws.String(key),
//			},
//		},
//		AttributesToGet: attributesToGet,
//		TableName:       aws.String(tableName),
//	}
//
//	result, err := d.conn.GetItemWithContext(ctx, input)
//	if err != nil {
//		return nil, errs.NewError(
//			errs.ErrInternal,
//			fmt.Errorf("error getting session: %v", err),
//		)
//	}
//
//	if len(result.Item) == 0 {
//		return nil, errs.NewError(errs.ErrNotFound, nil)
//	}
//
//	var values map[string]interface{}
//	err = dynamodbattribute.UnmarshalMap(result.Item, &values)
//	if err != nil {
//		return nil, errs.NewError(
//			errs.ErrInternal,
//			fmt.Errorf("error unmarshalling session: %v", err),
//		)
//	}
//	return values, nil
//}
//
//func (d dynamoRepository) GetTTL(ctx context.Context, tableName, key string) (time.Time, errs.ChatError) {
//	result, err := d.HashGet(ctx, tableName, key, "ttl")
//	ttl := result["ttl"].(float64)
//	if err != nil {
//		return time.Time{}, err
//	}
//
//	return time.Unix(int64(ttl), 0), nil
//}
//
//func (d dynamoRepository) HashGetEncrypted(
//	ctx context.Context,
//	tableName, key, secret string,
//) (map[string]interface{}, errs.ChatError) {
//	result, err := d.HashGet(ctx, tableName, key, "encrypted_value")
//	if err != nil {
//		return nil, errs.NewError(
//			errs.ErrInternal,
//			fmt.Errorf("error getting session: %v", err),
//		)
//	}
//	decryptedValue, errDecrypt := d.encryptor.Decrypt(result["encrypted_value"].(string), secret)
//	if errDecrypt != nil {
//		return nil, errs.NewError(
//			errs.ErrInternal,
//			fmt.Errorf("error decrypting value: %v", err),
//		)
//	}
//
//	var value map[string]interface{}
//	errUnmarshall := json.Unmarshal([]byte(decryptedValue), &value)
//	if errUnmarshall != nil {
//		return nil, errs.NewError(
//			errs.ErrInternal,
//			fmt.Errorf("corrupted data"),
//		)
//	}
//
//	return value, nil
//}
//
//func (d dynamoRepository) ExpireAt(ctx context.Context, tableName string, key string, at time.Time) errs.ChatError {
//	input := &dynamodb.UpdateItemInput{
//		UpdateExpression: aws.String("SET #ttl = :ttl"),
//		ExpressionAttributeNames: map[string]*string{
//			"#ttl": aws.String("ttl"),
//		},
//		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
//			":ttl": {
//				N: aws.String(strconv.FormatInt(at.Unix(), 10)),
//			},
//		},
//		Key: map[string]*dynamodb.AttributeValue{
//			"id": {
//				S: aws.String(key),
//			},
//		},
//		TableName: aws.String(tableName),
//	}
//	_, err := d.conn.UpdateItemWithContext(ctx, input)
//	if err != nil {
//		return errs.NewError(
//			errs.ErrInternal,
//			fmt.Errorf("error setting ttl: %v", err),
//		)
//	}
//	return nil
//}
//
//func NewDynamodbRepository(conn *dynamodb.DynamoDB, encryptor encryptor.Encryptor) (sessionManager.Repository, error) {
//	err := createTableSession(conn)
//	if err != nil {
//		return nil, err
//	}
//	err = enableTTL(conn, "session")
//	if err != nil {
//		return nil, err
//	}
//	err = createTableUserSession(conn)
//	if err != nil {
//		return nil, err
//	}
//	err = enableTTL(conn, "user_session")
//	if err != nil {
//		return nil, err
//	}
//
//	return &dynamoRepository{conn: conn, encryptor: encryptor}, nil
//}
//
//func createTableSession(conn *dynamodb.DynamoDB) error {
//	tableName := "session"
//
//	input := &dynamodb.CreateTableInput{
//		AttributeDefinitions: []*dynamodb.AttributeDefinition{
//			{
//				AttributeName: aws.String("id"),
//				AttributeType: aws.String("S"),
//			},
//		},
//		KeySchema: []*dynamodb.KeySchemaElement{
//			{
//				AttributeName: aws.String("id"),
//				KeyType:       aws.String("HASH"),
//			},
//		},
//		TableName:   aws.String(tableName),
//		BillingMode: aws.String(dynamodb.BillingModePayPerRequest),
//	}
//
//	_, err := conn.CreateTable(input)
//	if err != nil && err.Error() != "ResourceInUseException: Table already exists: session" {
//		return err
//	}
//
//	return waitTableCreate(conn, tableName)
//}
//
//func createTableUserSession(conn *dynamodb.DynamoDB) error {
//	tableName := "user_session"
//
//	input := &dynamodb.CreateTableInput{
//		AttributeDefinitions: []*dynamodb.AttributeDefinition{
//			{
//				AttributeName: aws.String("id"),
//				AttributeType: aws.String("S"),
//			},
//			{
//				AttributeName: aws.String("session_id"),
//				AttributeType: aws.String("S"),
//			},
//		},
//		KeySchema: []*dynamodb.KeySchemaElement{
//			{
//				AttributeName: aws.String("id"),
//				KeyType:       aws.String("HASH"),
//			},
//			{
//				AttributeName: aws.String("session_id"),
//				KeyType:       aws.String("RANGE"),
//			},
//		},
//		TableName:   aws.String(tableName),
//		BillingMode: aws.String(dynamodb.BillingModePayPerRequest),
//	}
//
//	_, err := conn.CreateTable(input)
//	if err != nil && err.Error() != "ResourceInUseException: Table already exists: user_session" {
//		return err
//	}
//
//	return waitTableCreate(conn, tableName)
//}
//
//func waitTableCreate(conn *dynamodb.DynamoDB, tableName string) error {
//	var status string
//	for {
//		time.Sleep(1 * time.Second)
//		currentTable, err := conn.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(tableName)})
//		if err != nil {
//			return err
//		}
//		status = *currentTable.Table.TableStatus
//		if status != dynamodb.TableStatusCreating {
//			break
//		}
//	}
//
//	if status != dynamodb.TableStatusActive {
//		return fmt.Errorf("table %s is not active", tableName)
//	}
//
//	return nil
//}
//
//func enableTTL(conn *dynamodb.DynamoDB, tableName string) error {
//	input := &dynamodb.UpdateTimeToLiveInput{
//		TableName: aws.String(tableName),
//		TimeToLiveSpecification: &dynamodb.TimeToLiveSpecification{
//			AttributeName: aws.String("ttl"),
//			Enabled:       aws.Bool(true),
//		},
//	}
//
//	_, err := conn.UpdateTimeToLive(input)
//	if err != nil && err.(awserr.Error).Message() != "TimeToLive is already enabled" {
//		return err
//	}
//	return nil
//}
