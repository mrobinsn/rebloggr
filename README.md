# rebloggr

Utility to transfer all of your posts from one Tumblr blog to another via reblogging.

**WARNING**: *this tool deletes the posts from the original blog after reblogging!*

## Install with
Download one of the pre-built binaries for your platform on the Releases tab.

Alternatively install from source (Requires a working [Go](https://golang.org/) installation):
```
go get -u github.com/mrobinsn/rebloggr
```

## Use

You must [register](https://www.tumblr.com/oauth/apps) an "application" with Tumblr to get a Consumer Key and Consumer Secret.

Setup environment (or optionally use CLI flags):
```
export REBLOGGR_CONSUMER_KEY=HKetaxxxxxxxxxp8ZTJVE
export REBLOGGR_CONSUMER_SECRET=zb4twxxxxxxxxxwPZoLB
export REBLOGGR_CALLBACK_URL=https://github.com/mrobinsn/rebloggr
```

Generate an OAUTH 1.0 token so that the tool can act on behalf of your Tumblr account:
```
$ rebloggr token
(1) Go to: https://www.tumblr.com/oauth/authorize?oauth_token=hQoSxxxxxktWH3
(2) Grant access, you should be redirected to a page with a "oauth_verifier" value in the URL.
(3) Enter that verification code here:
uayQxxxxxxxxxxxxxxxxx5z

Token: LGwduayQxxxxxxxxxxxxxxxxx5z4QCx
Secret: jFjuayQxxxxxxxxxxxxxxxxx5z6UQZ

Token written to .token
```

This will write a `.token` file to your current directory with a valid OAUTH token.
You should delete this file after you are finished with the application.

Now that you have a token, you can run the reblog command:

```
$ rebloggr reblog
Hello, the-miniaturestudentangel!
Looks like you have 2 blog(s)

Which blog to reblog FROM?

1. the-miniaturestudentangel.tumblr.com
2. testblog12515.tumblr.com

Enter a number: 1

Which blog to post TO?

1. testblog12515.tumblr.com

Enter a number: 1

Preparing to reblog everything from the-miniaturestudentangel.tumblr.com to testblog12515.tumblr.com
!! THIS WILL DELETE POSTS FROM the-miniaturestudentangel.tumblr.com AFTER REBLOGGING TO testblog12515.tumblr.com !!
Are you sure you want to continue? (y/n) [n]: y
Reblogging..
[1] Reblogged 145132930378 - [the-url] to testblog12515.tumblr.com
[1] Deleted 145132930378 - [the-url]from the-miniaturestudentangel.tumblr.com
[2] Reblogged 145132922278 - [the-url] to testblog12515.tumblr.com
[2] Deleted 145132922278 - [the-url] from the-miniaturestudentangel.tumblr.com
[3] Reblogged 145132920193 - [the-url] to testblog12515.tumblr.com
[3] Deleted 145132920193 - [the-url] from the-miniaturestudentangel.tumblr.com
[4] Reblogged 145132918238 - [the-url] to testblog12515.tumblr.com
[4] Deleted 145132918238 - [the-url] from the-miniaturestudentangel.tumblr.com
[5] Reblogged 145132916193 - [the-url] to testblog12515.tumblr.com
[5] Deleted 145132916193 - [the-url] from the-miniaturestudentangel.tumblr.com
[6] Reblogged 145132911668 - [the-url] to testblog12515.tumblr.com
[6] Deleted 145132911668 - [the-url] from the-miniaturestudentangel.tumblr.com
[7] Reblogged 145132906568 - [the-url] to testblog12515.tumblr.com
[7] Deleted 145132906568 - [the-url] from the-miniaturestudentangel.tumblr.com
[8] Reblogged 145132903513 - [the-url] to testblog12515.tumblr.com
[8] Deleted 145132903513 - [the-url] from the-miniaturestudentangel.tumblr.com
Reblogged 8 post(s)!
```

It's that easy!

Keep in mind that Tumblr enforces an account limit of 250 posts per *day*. So depending on how many posts you have to move, it may take multiple days. If this happens you will see the following error output:
```
You have hit the limit of 250 posts per day, try running again after 24hrs.
```

Since the tool deletes each post after it re-blogs it, you can just run the tool again the next day and it will pick back up where it left off.

## Contributing

Currently, I've smushed everything into one `main.go` file, and it could use some clean-up/organization if anyone feels up to the task. Specifically the `reblog` function has a high cyclomatic complexity due to its length.

Pull requests are welcome. I built this tool to help my wife move ~1200 posts across her blogs, but I went ahead and open sourced it in the hopes that my work will help someone else.
