/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package authtls

import (
	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/apis/extensions"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	ing_errors "k8s.io/ingress/core/pkg/ingress/errors"
	"k8s.io/ingress/core/pkg/ingress/resolver"
	"k8s.io/ingress/core/pkg/k8s"
)

const (
	// name of the secret used for authentication
	annotationAuthTLSSecret = "ingress.kubernetes.io/auth-tls-secret"
	// Verify Depth for cert authentication
	annotationAuthTLSDepth = "ingress.kubernetes.io/auth-tls-verify-depth"
	defaultAuthTLSDepth    = 1
	// Error page to redirect when authentication fails
	annotationAuthTLSErrorpage = "ingress.kubernetes.io/auth-tls-error-page"
)

// AuthSSLConfig contains the AuthSSLCert used for muthual autentication
// and the configured ValidationDepth
type AuthSSLConfig struct {
	AuthSSLCert     resolver.AuthSSLCert
	ValidationDepth int    `json:"validationDepth"`
	AuthErrorPage   string `json:"errorPage"`
}

// NewParser creates a new TLS authentication annotation parser
func NewParser(resolver resolver.AuthCertificate) parser.IngressAnnotation {
	return authTLS{resolver}
}

type authTLS struct {
	certResolver resolver.AuthCertificate
}

// Parse parses the annotations contained in the ingress
// rule used to use a Certificate as authentication method
func (a authTLS) Parse(ing *extensions.Ingress) (interface{}, error) {

	tlsauthsecret, err := parser.GetStringAnnotation(annotationAuthTLSSecret, ing)
	if err != nil {
		return &AuthSSLConfig{}, err
	}

	if tlsauthsecret == "" {
		return &AuthSSLConfig{}, ing_errors.NewLocationDenied("an empty string is not a valid secret name")
	}

	_, _, err = k8s.ParseNameNS(tlsauthsecret)
	if err != nil {
		return &AuthSSLConfig{}, ing_errors.NewLocationDenied("an empty string is not a valid secret name")
	}

	tlsdepth, err := parser.GetIntAnnotation(annotationAuthTLSDepth, ing)
	if err != nil || tlsdepth == 0 {
		tlsdepth = defaultAuthTLSDepth
	}

	authCert, err := a.certResolver.GetAuthCertificate(tlsauthsecret)
	if err != nil {
		return &AuthSSLConfig{}, ing_errors.LocationDenied{
			Reason: errors.Wrap(err, "error obtaining certificate"),
		}
	}

	authTLSErrorPage, err := parser.GetStringAnnotation(annotationAuthTLSSecret, ing)
	if err != nil {
		return &AuthSSLConfig{}, err
	}

	return &AuthSSLConfig{
		AuthSSLCert:     *authCert,
		ValidationDepth: tlsdepth,
		AuthErrorPage:   authTLSErrorPage,
	}, nil
}
