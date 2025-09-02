package authentication

import (
	"context"
	"log"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

func NewSpotifyClient(ctx context.Context, tok *oauth2.Token, authenticator *spotifyauth.Authenticator) *spotify.Client {
	if tok == nil {
		log.Println("NewSpotifyClient: nil oauth2.Token")
		return nil
	}

	if authenticator == nil {
		log.Println("NewSpotifyClient: nil spotifyauth.Authenticator")
		return nil
	}

	return spotify.New(authenticator.Client(ctx, tok), spotify.WithRetry(true))
}
