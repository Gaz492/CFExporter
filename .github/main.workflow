workflow "Build Application" {
  on = "push"
  resolves = ["release"]
}

action "build" {
  uses = "gaz492/actions/golang-build@master"
  args = "windows/amd64 windows/386 linux/amd64 darwin/amd64"
}

action "release" {
  needs = ["build"]
  uses = "actions/aws/cli@master"
  args = "aws --endpoint-url https://s3.gaz492.uk s3 sync .release s3://tools"
  // Make sure to add the secrets in the repo settings page
  // AWS_REGION is set to us-east-1 by default
  secrets = [
    "AWS_ACCESS_KEY_ID",
    "AWS_SECRET_ACCESS_KEY",
  ]
}
