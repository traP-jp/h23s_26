package handler

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	traqoauth2 "github.com/ras0q/traq-oauth2"
	"github.com/traP-jp/h23s_26/internal/pkg/config"
)

type (
	AuthorizeRequest struct {
		CodeChallengeMethod string `query:"code_challenge_method"`
		State               string `query:"state"`
	}

	CallbackRequest struct {
		Code string `query:"code"`
	}
)

func (h *Handler) Authorize(c echo.Context) error {
	var req AuthorizeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body").SetInternal(err)
	}

	codeChallengeMethod, ok := traqoauth2.CodeChallengeMethodFromStr(req.CodeChallengeMethod)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid code_challenge_method")
	}

	codeVerifier, authURL, err := h.traqOauth2Config.AuthorizeWithPKCE(codeChallengeMethod, req.State)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	sess, _ := session.Get(config.SessionName, c)
	sess.Values[config.CodeVerifierKey] = codeVerifier
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	return c.Redirect(http.StatusFound, authURL)
}

func (h *Handler) Callback(c echo.Context) error {
	var req CallbackRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body").SetInternal(err)
	}

	if req.Code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	sess, err := session.Get(config.SessionName, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	codeVerifier, ok := sess.Values[config.CodeVerifierKey].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	tok, err := h.traqOauth2Config.CallbackWithPKCE(
		c.Request().Context(),
		codeVerifier,
		req.Code,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	sess.Values[config.TokenKey] = tok
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	// return c.String(http.StatusOK, "You are logged in!")
	return c.JSON(200, tok)
}
