Export the following in `.env` (or otherwise), replacing keys as appropriate:

```shell
export AWS_ACCESS_KEY="ABCDEFGHIJKL"
export AWS_SECRET_ACCESS_KEY="1234567890qwertyuiopasdfghjklzxcvbnm"

export SCREENSHOT_DIR="~/Dropbox/Screenshots"
export SCREENSHOT_BUCKET="i.samby.co.uk"
export SCREENSHOT_BUCKET_DIR="i"
```

and add something like this to your .bashrc/.zshrc for the s3 command:

```shell
# Setup screenshots
alias s3='sh -c "source ~/Dev/Go/src/github.com/sambooo/dropbox-s3/.env; dropbox-s3"'
```
