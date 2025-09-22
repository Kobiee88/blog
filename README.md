# blog

WARNING: Do NOT use any of this! It is not safe! ;)

Anyways ...
You need Postgres and Go installed, to make it work.
Run go install blog (yes, I named my project blog instead of gator for some reason) in the root directory.

The following commands are available:
- "login" - "[user]": Logs in user
- "register" - "[user]": Registers new user
- "reset": Resets the database
- "users": Returns registeres users
- "agg": Activates the gator
- "addfeed" - ["Feed"] - ["URL"]: Adds a feed to the gator
- "feeds": Returns all feeds
- "follow" - ["URL"]: Follows a feed
- "following": Returns followed feeds
- "unfollow" - ["URL"]: Unfollows feed
- "browse" - optional "[limit]": Returns a list of posts