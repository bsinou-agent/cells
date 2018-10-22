#!/bin/bash
# Generate a Markdown change log of pull requests from commits between two tags
# Forked from URL: https://gist.github.com/kingkool68/09a201a35c83e43af08fcbacee5c315a , Author: Russell Heimlich 

# Repo URL to base links off of
REPOSITORY_URL=https://github.com/pydio/cells

# Get a list of all tags in reverse order
# Assumes the tags are in version format like v1.2.3
GIT_TAGS=$(git tag -l --sort=-version:refname)

# Make the tags an array
TAGS=($GIT_TAGS)
#LATEST_TAG=${TAGS[0]}
#PREVIOUS_TAG=${TAGS[1]}
LATEST_TAG=v$VERSION
PREVIOUS_TAG=${TAGS[0]}

# TODO: Insure version is greater than previous tag
echo "Generating change log between "$PREVIOUS_TAG" and "$LATEST_TAG"..."

# Get a log of commits that occured between two tags
# We only get the commit hash so we don't have to deal with a bunch of ugly parsing
# See Pretty format placeholders at https://git-scm.com/docs/pretty-formats
# We retrieve logs until head: Latest tag hasn't been done yet.
COMMITS=$(git log $PREVIOUS_TAG..HEAD --pretty=format:"%H")

# Store our changelog in a variable to be saved to a file at the end
MARKDOWN="# Pydio Cells - Changes in $LATEST_TAG\n\n"
MARKDOWN+="[See Full Changelog]($REPOSITORY_URL/compare/$PREVIOUS_TAG...$LATEST_TAG)"
MARKDOWN+='\n'

HTML="<a href='$REPOSITORY_URL/compare/$PREVIOUS_TAG...$LATEST_TAG'>$LATEST_TAG - full Changelog</a>"
HTML+='\n'
HTML+="<ul>"

# Loop over each commit and look for merged pull requests
for COMMIT in $COMMITS; do
	# Get the subject of the current commit
	SUBJECT=$(git log -1 ${COMMIT} --pretty=format:"%s")

	# If the subject contains "Merge pull request #xxxxx" then it is deemed a pull request
	PULL_REQUEST=$( grep -Eo "Merge pull request #[[:digit:]]+" <<< "$SUBJECT" )
	MERGE_BRANCH=$( grep -Eo "Merge branch" <<< "$SUBJECT" )
	VENDOR_COMMIT=$( grep -Eoi "vendor" <<< "$SUBJECT" )
	if [[ $PULL_REQUEST ]]; then
		# Perform a substring operation so we're left with just the digits of the pull request
		PULL_NUM=${PULL_REQUEST#"Merge pull request #"}
		# AUTHOR_NAME=$(git log -1 ${COMMIT} --pretty=format:"%an")
		# AUTHOR_EMAIL=$(git log -1 ${COMMIT} --pretty=format:"%ae")

		# Get the body of the commit
		BODY=$(git log -1 ${COMMIT} --pretty=format:"%b")
		MARKDOWN+='\n'
		HTML+='\n'
		MARKDOWN+="- [#$PULL_NUM]($REPOSITORY_URL/pull/$PULL_NUM): $BODY"
        HTML+="<li><a href="$REPOSITORY_URL/commit/$PULL_NUM">#${PULL_NUM}</a>: $BODY</li>"
    else
        if [ -z "$MERGE_BRANCH" ] && [ -z "$VENDOR_COMMIT" ]; then
            MARKDOWN+='\n'
    		HTML+='\n'
            MARKDOWN+="- [#${COMMIT:0:7}]($REPOSITORY_URL/commit/$COMMIT): $SUBJECT"
            HTML+="<li><a href="$REPOSITORY_URL/commit/$COMMIT">#${COMMIT:0:7}</a>: $SUBJECT</li>"
		fi
	fi
done

HTML+='\n'
HTML+='</ul>'

# Save our markdown to a file
echo -e $MARKDOWN > CHANGELOG.md
echo -e $HTML > CHANGELOG.html
