package routes

import (
	"time"

	"github.com/go-macaron/session"
	"github.com/hoffx/infoimadvent/storage"
	macaron "gopkg.in/macaron.v1"
)

func Home(ctx *macaron.Context, sess session.Store) {
	value := sess.Get("user")
	sUser, ok := value.(storage.User)
	if ok && sUser.Active {
		ctx.Data["LoggedIn"] = true
	}
	ctx.Data["IsAdvent"] = isAdvent()
	ctx.HTML(200, "home")
}

func isAdvent() bool {
	_, m, d := time.Now().Date()
	if m == time.January && d >= 1 && d <= 24 {
		return true
	}
	return false
}
