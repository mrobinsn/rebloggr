package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/MariaTerzieva/gotumblr"
	log "github.com/Sirupsen/logrus"
	"github.com/Songmu/prompter"
	"github.com/fatih/color"
	"github.com/mrjones/oauth"
	"github.com/tcnksm/go-input"

	"github.com/codegangsta/cli"
)

const (
	version = "0.0.1"
	name    = "rebloggr"

	endpoint = "https://api.tumblr.com"
)

var (
	app = initApp()

	consumer                                 *oauth.Consumer
	consumerKey, consumerSecret, callbackURL string
)

func initApp() *cli.App {
	newApp := cli.NewApp()

	newApp.Name = name
	newApp.Version = version
	newApp.Usage = "utility to reblog all posts from one blog to another"
	newApp.Authors = []cli.Author{
		{Name: "Michael Robinson", Email: "mrobinson@outlook.com"},
	}
	newApp.Before = globalSetup

	// global flags
	newApp.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level",
			Usage: "log level",
			Value: "INFO",
		},
		cli.StringFlag{
			Name:   "consumer-key",
			EnvVar: "REBLOGGR_CONSUMER_KEY",
			Usage:  "Consumer Key generated during application registration",
			Value:  "",
		},
		cli.StringFlag{
			Name:   "consumer-secret",
			EnvVar: "REBLOGGR_CONSUMER_SECRET",
			Usage:  "Consumer Secret generated during application registration",
			Value:  "",
		},
		cli.StringFlag{
			Name:   "callback-url",
			EnvVar: "REBLOGGR_CALLBACK_URL",
			Usage:  "Consumer URL setup during application registration",
			Value:  "",
		},
	}

	newApp.Commands = append(newApp.Commands,
		cli.Command{
			Name:   "token",
			Usage:  "get a OAUTH token",
			Action: token,
		},
		cli.Command{
			Name:   "reblog",
			Usage:  "start the reblogging process",
			Action: reblog,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "token-key",
					EnvVar: "REBLOGGR_TOKEN_KEY",
					Usage:  "OAUTH token key",
					Value:  "",
				},
				cli.StringFlag{
					Name:   "token-secret",
					EnvVar: "REBLOGGR_TOKEN_SECRET",
					Usage:  "OAUTH token secret key",
					Value:  "",
				},
			},
		},
	)

	return newApp
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Fatal("app failed")
	}
}

func globalSetup(c *cli.Context) error {
	configureLogging(c.String("log-level"))

	consumerKey = c.String("consumer-key")
	consumerSecret = c.String("consumer-secret")
	callbackURL = c.String("callback-url")

	if consumerKey == "" {
		return fmt.Errorf("consumer key is required")
	}
	if consumerSecret == "" {
		return fmt.Errorf("consumer secret is required")
	}
	if callbackURL == "" {
		return fmt.Errorf("callback url is required")
	}

	consumer = oauth.NewConsumer(
		consumerKey,
		consumerSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   "https://www.tumblr.com/oauth/request_token",
			AuthorizeTokenUrl: "https://www.tumblr.com/oauth/authorize",
			AccessTokenUrl:    "https://www.tumblr.com/oauth/access_token",
		})
	return nil
}

func configureLogging(level string) {
	switch strings.ToLower(level) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}
}

func token(c *cli.Context) error {
	requestToken, u, err := consumer.GetRequestTokenAndUrl(callbackURL)
	if err != nil {
		return err
	}

	fmt.Println("(1) Go to: " + color.RedString(u))
	fmt.Println("(2) Grant access, you should be redirected to a page with a \"oauth_verifier\" value in the URL.")
	fmt.Println("(3) Enter that verification code here: ")

	verificationCode := ""
	fmt.Scanln(&verificationCode)

	accessToken, err := consumer.AuthorizeToken(requestToken, verificationCode)
	if err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("Token: " + color.GreenString(accessToken.Token))
	fmt.Println("Secret: " + color.GreenString(accessToken.Secret))
	fmt.Println("")

	// Write the token to a file
	jsonToken, err := json.Marshal(accessToken)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(".token", jsonToken, 0660); err != nil {
		return err
	}
	color.Cyan("Token written to .token")
	return nil
}

// FIXME: cyclomatic complexity is high, refactor into multiple logical chunks
func reblog(c *cli.Context) error {
	tokenKey := c.String("token-key")
	tokenSecret := c.String("token-secret")

	token := oauth.AccessToken{Token: tokenKey, Secret: tokenSecret}
	if tokenKey == "" || tokenSecret == "" {
		// Try to get the token from the token file
		file, err := os.Open(".token")
		if err == os.ErrNotExist {
			return fmt.Errorf("token must be provided, run the `token` command first")
		}
		if err != nil {
			return err
		}

		if err := json.NewDecoder(file).Decode(&token); err != nil {
			return err
		}
		_ = file.Close()
	}

	client := gotumblr.NewTumblrRestClient(consumerKey, consumerSecret, token.Token, token.Secret, callbackURL, endpoint)

	// Get the user's info
	userInfo := client.Info()
	fmt.Printf("Hello, %s!\n", color.GreenString(userInfo.User.Name))
	blogs := userInfo.User.Blogs
	fmt.Printf("Looks like you have %s blog(s)\n", color.GreenString(strconv.Itoa(len(blogs))))
	fmt.Println("")

	if len(blogs) < 2 {
		return fmt.Errorf("You must have at least 2 blogs to use this tool..")
	}

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	// Choose a source blog
	sourceChoices := make([]string, len(blogs))
	for i, blog := range blogs {
		sourceChoices[i] = hostOnly(blog.Url)
	}
	source, err := ui.Select(fmt.Sprintf("Which blog to reblog %s?", color.RedString("FROM")), sourceChoices, &input.Options{Required: true, Loop: true})
	if err != nil {
		return err
	}

	destChoices := make([]string, 0, len(blogs)-1)
	for _, blog := range blogs {
		if hostOnly(blog.Url) != source {
			destChoices = append(destChoices, hostOnly(blog.Url))
		}
	}
	// Choose a destination blog
	dest, err := ui.Select(fmt.Sprintf("Which blog to post %s?", color.RedString("TO")), destChoices, &input.Options{Required: true, Loop: true})
	if err != nil {
		return err
	}

	fmt.Printf("Preparing to reblog everything from %s to %s\n", color.CyanString(source), color.GreenString(dest))
	fmt.Printf("!! %s %s %s %s !!\n", color.RedString("THIS WILL DELETE POSTS FROM"), color.CyanString(source), color.RedString("AFTER REBLOGGING TO"), color.CyanString(dest))

	if !prompter.YN("Are you sure you want to continue?", false) {
		return nil
	}

	fmt.Println("Reblogging..")
	total := 0
	for {
		posts := client.Posts(source, "", map[string]string{"offset": "0"})
		if len(posts.Posts) == 0 {
			break
		}
		for _, rawPost := range posts.Posts {
			// Parse the post
			post := gotumblr.BasePost{}
			if err := json.Unmarshal(rawPost, &post); err != nil {
				return err
			}

			// Repost it
			if err := client.Reblog(dest, map[string]string{"id": fmt.Sprintf("%d", post.Id), "reblog_key": post.Reblog_key}); err != nil {
				if err.Error() == "Bad Request" {
					color.Yellow("You have hit the limit of 250 posts per day, try running again after 24hrs.")
					return nil
				}
				return err
			}
			total++
			fmt.Printf("[%s] Reblogged %d - %s to %s\n", color.GreenString("%d", total), post.Id, color.CyanString(post.Post_url), color.CyanString(dest))

			// Delete it
			if err := client.DeletePost(source, fmt.Sprintf("%d", post.Id)); err != nil {
				return err
			}
			fmt.Printf("[%s] Deleted %d - %s from %s\n", color.RedString("%d", total), post.Id, color.CyanString(post.Post_url), color.CyanString(source))
		}
	}

	fmt.Printf("Reblogged %s post(s)!\n", color.GreenString("%d", total))
	return nil
}

func hostOnly(aURL string) string {
	pURL, err := url.Parse(aURL)
	if err != nil {
		panic(err)
	}
	return pURL.Host
}
