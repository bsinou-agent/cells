#!/bin/bash

# call
# env OAUTH_TOKEN=yourtoken ./prerelease.sh

# Default Values
DFLT_REPO_OWNER=pydio
DFLT_REPO_ID=cells
DFLT_FROM_BRANCH=master

#OAUTH_TOKEN=atokenthatshouldwork....

echo ""
echo "Welcome to Cells Release procedure."
echo "We will guide you trough the process and ask you a few questions... Get ready!"
echo ""

## Retrieve git coordinates for the code to release

#read -p "   Please enter the git user name that will be used: " GIT_USER
#if [ "x`printf '%s' "$GIT_USER" | tr -d "$IFS"`" = x ]; then
#    echo "You should provide a valid user name."
#    echo "Aborting process"
#    exit 1
#fi 

read -p "   Please enter repository owner. (default: $DFLT_REPO_OWNER): " REPO_OWNER
if [ "x`printf '%s' "$REPO_OWNER" | tr -d "$IFS"`" = x ]; then
    REPO_OWNER=$DFLT_REPO_OWNER
fi

read -p "   Please enter repository ID. (default: $DFLT_REPO_ID): " REPO_ID
if [ "x`printf '%s' "$REPO_ID" | tr -d "$IFS"`" = x ]; then
    REPO_ID=$DFLT_REPO_ID
fi

#GIT_URL="https://$GIT_USER@github.com/$REPO_OWNER/$REPO_ID"
GIT_URL="https://github.com/$REPO_OWNER/$REPO_ID"
API_URL="https://api.github.com"

## Retrieve release info

read -p "   Which branch do you want to use as source (default: $DFLT_FROM_BRANCH): " FROM_BRANCH
if [ "x`printf '%s' "$FROM_BRANCH" | tr -d "$IFS"`" = x ]; then
    FROM_BRANCH=$DFLT_FROM_BRANCH
fi 

# Get a list of all tags in reverse order
# Assumes the tags are in version format like v1.2.3
GIT_TAGS=$(git tag -l --sort=-version:refname)
# Make the tags an array
TAGS=($GIT_TAGS)
PREVIOUS_TAG=${TAGS[0]}

last_version=${PREVIOUS_TAG:1}

major=`echo $last_version | cut -d. -f1`
minor=`echo $last_version | cut -d. -f2`
revision=`echo $last_version | cut -d. -f3`
revision=`expr $revision + 1`

new_version="$major.$minor.$revision"

read -p "   Last version is $last_version, which version do you want to release now (default: "$new_version"): " u_version
if ! [ "x`printf '%s' "$u_version" | tr -d "$IFS"`" = x ]; then
    # TODO validate format
    new_version=$u_version
fi 
NEW_TAG=v$new_version

VERSION=$new_version
TMP_BRANCH=pre-release-$VERSION

read -p "   Please provide a short description for this release: " SHORT_DESC
if [ "x`printf '%s' "$SHORT_DESC" | tr -d "$IFS"`" = x ]; then
    echo "Description cannot be empty."
    echo "Aborting process"
    exit 1
fi 

# TODO add a few more checks
# - check if GIT_URL == origin
# - check if given version is valid
# - check if we have read write access on the repo
# - check if tag and release alrewady exist
# ...

date
echo "About to release $VERSION on $GIT_URL: $SHORT_DESC"

## Prepare release in a local branch
git fetch origin
git checkout stable
git checkout -b $TMP_BRANCH
git merge $FROM_BRANCH
make clean generate main CELLS_VERSION=$VERSION
go test ./...
env VERSION=$VERSION ./changelog.sh
git add -A
git commit -am "Release $NEW_TAG"
git tag -a $NEW_TAG -m "$NEW_TAG"

## Confirm and publish on github
createReleaseUrl="$API_URL/repos/$REPO_OWNER/$REPO_ID/releases"

read -p "You are about to push your modifications. Are you sure you want to proceed? (y/n, default is yes): " choice
case "$choice" in 
  y|Y|'' ) ;;
  n|N|* ) 
    echo "Aborting, nothing has been pushed to origin."; 
    echo "You should revert local change by issuing following commands:"
    echo "git tag -d $NEW_TAG ## <= Remove tag"
    echo "git checkout master"
    echo "git branch -D $TMP_BRANCH ## <= Delete tmp branch"
    exit 0
    ;;
esac

git push origin $TMP_BRANCH
git push origin $NEW_TAG

# Prepare and pre-publish release note on Github
md_changes="\n\n## Change log\n\nYou can find [a summary of the change log here]($GIT_URL/compare/$PREVIOUS_TAG...$NEW_TAG).\n"
json="{\"tag_name\": \"$NEW_TAG\",\"target_commitish\": \"$TMP_BRANCH\",\"name\": \"$NEW_TAG\",\"body\": \"$SHORT_DESC $md_changes\",\"draft\": false, \"prerelease\": true}"

echo "Creating release @ $createReleaseUrl with json: $json"
echo $json | curl -v -H "Authorization: token $OAUTH_TOKEN" --header "Content-Type: application/json" --request POST --data @- $createReleaseUrl

echo "Pre-release done."
date

exit 0

# Finalise release 
git pull origin $TMP_BRANCH
git checkout stable
git merge $TMP_BRANCH
git push origin stable
git push -d origin $TMP_BRANCH
git branch -d $TMP_BRANCH



