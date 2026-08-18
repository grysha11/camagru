package mailer

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var templates = template.Must(template.ParseFS(templateFS, "templates/*.tmpl"))

type linkEmailData struct {
	Intro    string
	LinkURL  string
	Footnote string
}

type commentNotificationData struct {
	CommenterUsername string
	CommentContent    string
	WallLinkURL       string
}

func RenderLinkEmail(intro, linkURL, footnote string) (string, error) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "link_email.html.tmpl", linkEmailData{
		Intro:    intro,
		LinkURL:  linkURL,
		Footnote: footnote,
	}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RenderCommentNotification(commenterUsername, commentContent, wallLinkURL string) (string, error) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "comment_notification.html.tmpl", commentNotificationData{
		CommenterUsername: commenterUsername,
		CommentContent:    commentContent,
		WallLinkURL:       wallLinkURL,
	}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
