#!/usr/bin/env bash

# 1. run tests before update
go test -bench=. -benchmem -benchtime=100000x > old.txt

# 2. update dependencies
go get -u ./... && go mod tidy && go get -u ./...

# 3. configure git
git config --global user.name 'update_deps robot'
git config --global user.email 'noreply@example.com'
git remote set-url origin https://x-access-token:${GH_TOKEN}@github.com/${REPO_NAME}
git checkout -b ${BRANCH_NAME}

# 4. commit changes or fast exit
git commit -am "Fix: $(date +%F) update dependencies" || exit 0
git push --set-upstream origin ${BRANCH_NAME} -f || exit 0

# 5. run tests after update
go test -bench=. -benchmem -benchtime=100000x > new.txt

# 6. create PR with benchmark difference
gh pr create -a ${REPO_OWNER} -b "$(benchcmp old.txt new.txt)" --fill || exit 0
