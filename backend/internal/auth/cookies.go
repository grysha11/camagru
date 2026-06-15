package auth

import "net/http"

func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name: "access_token",
		Value: accessToken,
		Path: "/",
		MaxAge: 15 * 60,
		HttpOnly: true,
		Secure: false,
		SameSite: http.SameSiteDefaultMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name: "refresh_token",
		Value: refreshToken,
		Path: "/",
		MaxAge: 7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure: false,
		SameSite: http.SameSiteDefaultMode,
	})
}
