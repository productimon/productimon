package main

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/local"
	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// Sign and verify auth tokens
type Authenticator struct {
	cert    *x509.Certificate
	certPEM []byte
	keyPEM  []byte
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
	signer  *local.Signer
}

// content of JWT claim
type Claims struct {
	Type string
	Uid  string
	Did  int64
	jwt.StandardClaims
}

// validity duration of a token
const TokenDuration = 1 * time.Hour

const (
	TokenVerifyType = "verify"
	TokenAuthType   = "auth"
)

// read or create root CA. This populates a.keyPEM and a.certPEM
func (a *Authenticator) initCert(certPath, keyPath string) error {
	var err error
	if a.certPEM, err = ioutil.ReadFile(certPath); err == nil {
		if a.keyPEM, err = ioutil.ReadFile(keyPath); err == nil {
			return nil
		}
	}
	log.Println("Initiating new certificate")

	mycsr := &csr.CertificateRequest{
		CN: "api.productimon.com",
		KeyRequest: &csr.KeyRequest{
			A: "rsa",
			S: 2048,
		},
		Names: []csr.Name{
			{
				C:  "AU",
				ST: "Sydney",
				L:  "Sydney",
				O:  "Productimon",
				OU: "Productimon mTLS",
			},
		},
	}

	if a.certPEM, _, a.keyPEM, err = initca.New(mycsr); err != nil {
		return err
	}

	if err = ioutil.WriteFile(certPath, a.certPEM, 0644); err != nil {
		return err
	}
	return ioutil.WriteFile(keyPath, a.keyPEM, 0600)
}

// Create a new Authenticator with given key pair location (create them if they don't exist)
func NewAuthenticator(publicKeyPath, privateKeyPath string) (*Authenticator, error) {
	var err error
	a := &Authenticator{}
	if err := a.initCert(publicKeyPath, privateKeyPath); err != nil {
		return nil, err
	}
	block, _ := pem.Decode(a.certPEM)
	if a.cert, err = x509.ParseCertificate(block.Bytes); err != nil {
		return nil, err
	}
	if a.privKey, err = jwt.ParseRSAPrivateKeyFromPEM(a.keyPEM); err != nil {
		return nil, err
	}
	if a.pubKey, err = jwt.ParseRSAPublicKeyFromPEM(a.certPEM); err != nil {
		return nil, err
	}
	if a.signer, err = local.NewSigner(a.privKey, a.cert, x509.SHA256WithRSA, nil); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Authenticator) GrpcCreds() (grpc.ServerOption, error) {
	serverCert, err := tls.X509KeyPair(a.certPEM, a.keyPEM)
	if err != nil {
		return grpc.EmptyServerOption{}, err
	}
	certPool := x509.NewCertPool()
	certPool.AddCert(a.cert)
	return grpc.Creds(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequestClientCert,
		ClientCAs:    certPool,
	})), nil

}

func (a *Authenticator) SignDeviceCert(uid string, did int64) (cert, key []byte, err error) {
	mycsr := &csr.CertificateRequest{
		CN: fmt.Sprintf("%d@%s.productimon.com", did, uid),
		KeyRequest: &csr.KeyRequest{
			A: "rsa",
			S: 2048,
		},
		Names: []csr.Name{
			{
				C:  a.cert.Subject.Country[0],
				ST: a.cert.Subject.Province[0],
				L:  a.cert.Subject.Locality[0],
				O:  a.cert.Subject.Organization[0],
				OU: a.cert.Subject.OrganizationalUnit[0],
			},
		},
	}
	csrPEM, key, err := csr.ParseRequest(mycsr)
	if err != nil {
		return nil, nil, err
	}
	req := signer.SignRequest{
		Hosts:    []string{},
		Request:  string(csrPEM),
		NotAfter: time.Now().Add(30 * 24 * time.Hour),
	}
	cert, err = a.signer.Sign(req)
	return
}

// Create a new JWT token for given uid
func (a *Authenticator) SignToken(uid string) (string, error) {
	expirationTime := time.Now().Add(TokenDuration)
	claims := Claims{
		Type: TokenAuthType,
		Uid:  uid,
		Did:  -1,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(a.privKey)
}

// Create a new verification token for given email
func (a *Authenticator) SignVerificationToken(email string) (string, error) {
	claims := Claims{
		Type: TokenVerifyType,
		Uid:  email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(a.privKey)
}

// Return uid for a given JWT token
func (a *Authenticator) VerifyToken(token string) (uid string, did int64, err error) {
	claims := &Claims{}

	p := jwt.Parser{ValidMethods: []string{jwt.SigningMethodRS256.Name}}
	tkn, err := p.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return a.pubKey, nil
	})

	if err != nil {
		return "", -1, err
	}

	if !tkn.Valid {
		return "", -1, errors.New("token is invalid")
	}

	if claims.Type != TokenAuthType {
		return "", -1, errors.New("invalid type")
	}

	return claims.Uid, claims.Did, nil
}

// Return email for a given JWT verification token
func (a *Authenticator) VerifyVerificationToken(token string) (email string, err error) {
	claims := &Claims{}

	p := jwt.Parser{ValidMethods: []string{jwt.SigningMethodRS256.Name}}
	tkn, err := p.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return a.pubKey, nil
	})

	if err != nil {
		return "", err
	}

	if !tkn.Valid {
		return "", errors.New("token is invalid")
	}

	if claims.Type != TokenVerifyType {
		return "", errors.New("invalid type")
	}

	return claims.Uid, nil
}

func (a *Authenticator) verifyCert(cert *x509.Certificate) (uid string, did int64, err error) {
	certPool := x509.NewCertPool()
	certPool.AddCert(a.cert)
	opts := x509.VerifyOptions{
		Roots: certPool,
	}
	_, err = cert.Verify(opts)
	if err != nil {
		return "", -1, err
	}
	log.Println(cert.Subject.CommonName)
	n, err := fmt.Sscanf(cert.Subject.CommonName, "%d@%s", &did, &uid)
	if err != nil {
		return "", -1, err
	}
	if n != 2 {
		return "", -1, errors.New("invalid certificate commonname")
	}
	uid = uid[0 : len(uid)-len(".productimon.com")]
	log.Printf("verified %s %d", uid, did)
	return
}

func (a *Authenticator) AuthenticateRequest(ctx context.Context) (uid string, did int64, err error) {
	peer, ok := peer.FromContext(ctx)
	if ok {
		tlsinfo, ok := peer.AuthInfo.(credentials.TLSInfo)
		if ok {
			certs := tlsinfo.State.PeerCertificates
			if len(certs) > 0 {
				uid, did, err = a.verifyCert(certs[0])
				if err == nil {
					log.Println(err)
					return
				}
			}
		}
	}

	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", -1, errors.New("metadata is not available")
	}
	auth := headers.Get("Authorization")
	if len(auth) != 1 {
		return "", -1, errors.New("authorization is missing")
	}
	return a.VerifyToken(auth[0])
}
