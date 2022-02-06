// Copyright 2022 Paul Greenberg greenpau@outlook.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authn

import (
	"context"
	"github.com/greenpau/go-authcrunch/pkg/requests"
	addrutil "github.com/greenpau/go-authcrunch/pkg/util/addr"
	"net/http"
	"net/url"
)

func (p *Portal) deleteAuthCookies(w http.ResponseWriter, r *http.Request) {
	for tokenName := range p.validator.GetAuthCookies() {
		w.Header().Add("Set-Cookie", p.cookie.GetDeleteCookie(addrutil.GetSourceHost(r), tokenName))
	}
}

func (p *Portal) handleHTTPLogout(ctx context.Context, w http.ResponseWriter, r *http.Request, rr *requests.Request) error {
	p.disableClientCache(w)
	p.injectRedirectURL(ctx, w, r, rr)
	h := addrutil.GetSourceHost(r)
	for tokenName := range p.validator.GetAuthCookies() {
		w.Header().Add("Set-Cookie", p.cookie.GetDeleteCookie(h, tokenName))
	}
	w.Header().Add("Set-Cookie", p.cookie.GetDeleteCookie(h, p.cookie.Referer))
	w.Header().Add("Set-Cookie", p.cookie.GetDeleteCookie(h, p.cookie.SessionID))
	return p.handleHTTPRedirect(ctx, w, r, rr, "/login")
}

func (p *Portal) handleHTTPLogoutWithLocalRedirect(ctx context.Context, w http.ResponseWriter, r *http.Request, rr *requests.Request) error {
	var refererExists bool
	p.disableClientCache(w)
	p.injectRedirectURL(ctx, w, r, rr)
	h := addrutil.GetSourceHost(r)
	for tokenName := range p.validator.GetAuthCookies() {
		w.Header().Add("Set-Cookie", p.cookie.GetDeleteCookie(h, tokenName))
	}
	if rr.Response.RedirectURL == "" {
		w.Header().Add("Set-Cookie", p.cookie.GetDeleteCookie(h, p.cookie.Referer))
	}
	w.Header().Add("Set-Cookie", p.cookie.GetDeleteCookie(h, p.cookie.SessionID))
	// The redirect_url query parameter exists.
	if rr.Response.RedirectURL != "" {
		return p.handleHTTPRedirect(ctx, w, r, rr, "/login?redirect_url="+rr.Response.RedirectURL)
	}
	// Find whether the redirect cookie exists. If so, do not inject redirect URL.
	if cookie, err := r.Cookie(p.cookie.Referer); err == nil {
		v, err := url.Parse(cookie.Value)
		if err == nil && v.String() != "" {
			refererExists = true
		}
	}
	if !refererExists {
		w.Header().Add("Set-Cookie", p.cookie.GetDeleteCookie(h, p.cookie.Referer))
		return p.handleHTTPRedirect(ctx, w, r, rr, "/login?redirect_url="+r.RequestURI)
	}
	return p.handleHTTPRedirect(ctx, w, r, rr, "/login")
}
