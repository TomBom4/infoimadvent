package routes

import (
	"encoding/base64"

	"github.com/go-macaron/session"
	"github.com/hoffx/infoimadvent/storage"
	macaron "gopkg.in/macaron.v1"
)

// Confirm handles the route "/confirm". The route is used by the confirmation
// link sent via mail.
func Confirm(ctx *macaron.Context, uStorer *storage.UserStorer, sess session.Store) {
	query := ctx.Req.URL.Query()
	var user storage.User
	// check if request is formatted correctly
	if query["user"] != nil && query["token"] != nil {
		// get according user
		var ok bool
		var element interface{}
		email, err := base64.StdEncoding.DecodeString(query["user"][0])
		if err != nil {
			ctx.Error(406, ctx.Tr(ErrWrongCredentials))
			ctx.Redirect("/", 406)
			return
		}
		element, err = uStorer.Get(map[string]interface{}{"email": string(email)})
		user, ok = element.(storage.User)
		if err != nil || !ok {
			ctx.Error(500, ctx.Tr(ErrDB))
			ctx.Redirect("/", 500)
			return
		}
		// check if token and user status are ok
		if user.Email != "" && !user.Confirmed && user.ConfirmationToken == query["token"][0] {
			user.ConfirmationToken = ""
			user.Confirmed = true
			user.Active = true

			sess.Set("user", user)

			err = uStorer.Put(user)
			if err != nil {
				ctx.Error(500, ctx.Tr(ErrDB))
				ctx.Redirect("/", 500)
				return
			}
		} else {
			// redirect user (that was messing around with the link) to home-page
			ctx.Error(406, ctx.Tr(ErrWrongCredentials))
			ctx.Redirect("/", 406)
			return
		}
	} else {
		ctx.Error(406, ctx.Tr(ErrWrongCredentials))
		ctx.Redirect("/", 406)
		return
	}
	ctx.Redirect("/login?Message="+ctx.Tr(MessLoggedIn), 302)
}
