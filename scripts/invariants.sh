#!/bin/sh
# Greppable invariants for the iiko clients. These are licence-seat and
# data-integrity rules, not style: breaking one affects a real restaurant, so
# they are checked mechanically rather than trusted to review.
set -u
fail=0

check() { # name, command, expected line count
	printf '%-52s' "$1"
	n=$(eval "$2" 2>/dev/null | wc -l | tr -d ' ')
	if [ "$n" -eq "$3" ]; then
		echo "ok"
	else
		echo "FAIL (got $n, want $3)"
		fail=1
	fi
}

# A check that passes because the thing it guards vanished is worse than one
# that fails, so first prove the trees are still here.
check "iikoserver packages present" "ls -d iikoserver/*/ | grep -c ." 1
check "iikocloud packages present"  "ls -d iikocloud/*/  | grep -c ." 1

check "no inline paths outside endpoints.go" \
	"grep -rn '\"/api/' iikoserver/*/*.go | grep -v endpoints.go | grep -v _test" 0

check "no inline date layouts outside dates.go" \
	"grep -rn '\"2006-' iikoserver iikocloud --include='*.go' | grep -v dates.go | grep -v _test" 0

check "rest does not import MCP" \
	"go list -deps ./iikoserver/rest | grep modelcontextprotocol" 0

# crypto/internal/entropy/v1.0.0 is stdlib whose path merely looks versioned.
check "rest has no non-stdlib dependency" \
	"go list -deps ./iikoserver/rest | grep -v '^vendor/' | grep -v '^crypto/internal' | grep '\\.' | grep -v iiko-go" 0

# The library packages must stay importable without pulling anything in.
# vendor/golang.org/x/... is Go's own vendored stdlib (crypto/tls, net/http
# pull it); it ships with the toolchain and is not a module dependency.
check "iikoserver has no non-stdlib dependency" \
	"go list -deps ./iikoserver/... | grep -v '^vendor/' | grep -v '^crypto/internal' | grep '\\.' | grep -v iiko-go" 0
check "iikocloud has no non-stdlib dependency" \
	"go list -deps ./iikocloud/... | grep -v '^vendor/' | grep -v '^crypto/internal' | grep '\\.' | grep -v iiko-go" 0

# The structural guarantee behind "every v2 JSON goes through DecodeV2": there
# is no plain JSON transport helper to bypass it with.
check "no JSON transport helper bypassing DecodeV2" \
	"grep -rn 'func (c \\*Client) \\(Get\\|Post\\)JSON' iikoserver" 0

check "licence-seat test present" \
	"grep -rn 'func TestReauthOn401AndLogoutReleasesSlot' iikoserver/rest" 1

check "sequential-request test present" \
	"grep -rn 'func TestRequestsAreSerialized' iikoserver/rest" 1

check "generated files are guarded against hand edits" \
	"grep -rn 'func TestGeneratedFilesMatchTheCatalog' tools/gen/server" 1

# No exported method may return an unexported type: the importer could call it
# but never name the result. This is why RawReport is exported.
check "no exported method returns an unexported type" \
	"go doc ./iikoserver/reports Service | grep -E '^func \\(s \\*Service\\)' | grep -E '\\) \\(\\*[a-z]'" 0

exit $fail
