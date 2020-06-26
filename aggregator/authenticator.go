package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"time"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/metadata"
)

// Sign and verify auth tokens
type Authenticator struct {
	signKey   *rsa.PrivateKey
	verifyKey *rsa.PublicKey
}

// content of JWT claim
type Claims struct {
	Uid string
	jwt.StandardClaims
}

// validity duration of a token
const TokenDuration = 1 * time.Hour

// Create a new Authenticator with given JWT key pair
func NewAuthenticator(publicKeyPath, privateKeyPath string) (*Authenticator, error) {
	JwtPubKey, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}
	JwtKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	ret := &Authenticator{}
	ret.signKey, err = jwt.ParseRSAPrivateKeyFromPEM(JwtKey)
	if err != nil {
		return nil, err
	}
	ret.verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(JwtPubKey)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Create a new JWT token for given uid
func (a *Authenticator) SignToken(uid string) (string, error) {
	expirationTime := time.Now().Add(TokenDuration)
	claims := Claims{
		Uid: uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(a.signKey)
}

// Return uid for a given JWT token
func (a *Authenticator) VerifyToken(token string) (uid string, err error) {
	claims := &Claims{}

	p := jwt.Parser{ValidMethods: []string{jwt.SigningMethodRS256.Name}}
	tkn, err := p.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return a.verifyKey, nil
	})

	if err != nil {
		return "", err
	}

	if !tkn.Valid {
		return "", errors.New("token is invalid")
	}

	return claims.Uid, nil
}

func (a *Authenticator) AuthenticateRequest(ctx context.Context) (uid string, err error) {
	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata is not available")
	}
	auth := headers.Get("Authorization")
	if len(auth) != 1 {
		return "", errors.New("authorization is missing")
	}
	return a.VerifyToken(auth[0])
}
