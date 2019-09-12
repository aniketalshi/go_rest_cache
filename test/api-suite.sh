#!/bin/bash

APP_PORT=${1:-7101}
HEALTHCHECK_PORT=${2:-$APP_PORT}
BASE_URL="http://localhost:$APP_PORT"
HEALTHCHECK_URL="http://localhost:$HEALTHCHECK_PORT"

for TOOL in bc curl jq wc awk sort uniq tr head tail; do
    if ! which $TOOL >/dev/null; then
        echo "ERROR: $TOOL is not available in the PATH"
        exit 1
    fi
done

PASS=0
FAIL=0
TOTAL=0

function describe() {
    echo -n "$1"; let TOTAL=$TOTAL+1
}

function pass() {
    echo "pass"; let PASS=$PASS+1
}

function fail() {
    RESPONSE=$1
    EXPECTED=$2
    echo "fail [$RESPONSE != $EXPECTED]";  let FAIL=$FAIL+1
}

function report() {
    PCT=$(echo "scale=2; $PASS / $TOTAL * 100" |bc)
    echo "$PASS/$TOTAL ($PCT%) tests passed"
}

describe "test-01-01: healthcheck = "

ATTEMPTS=0
while true; do
    let ATTEMPTS=$ATTEMPTS+1
    RESPONSE=$(curl -s -o /dev/null -w '%{http_code}' "$HEALTHCHECK_URL/healthcheck")
    if [[ $RESPONSE == "200" ]]; then
        let TIME=$ATTEMPTS*15
        echo -n "($TIME seconds) "; pass
        break
    else
        if [[ $ATTEMPTS -gt 24 ]]; then
            let TIME=$ATTEMPTS*15
            echo -n "($TIME seconds) "; fail
            break
        fi
        sleep 15
    fi
done

describe "test-02-01: / key count = "

COUNT=$(curl -s "$BASE_URL" |jq -r 'keys |.[]' |wc -l |awk '{print $1}')

if [[ $COUNT -eq 31 ]]; then
    pass
else
    fail "$COUNT" "31"
fi

describe "test-02-02: / repository_search_url value = "

VALUE=$(curl -s "$BASE_URL" |jq -r '.repository_search_url')

if [[ "$VALUE" == "https://api.github.com/search/repositories?q={query}{&page,per_page,sort,order}" ]]; then
    pass
else
    fail "$VALUE" "https://api.github.com/search/repositories?q={query}{&page,per_page,sort,order}"
fi

describe "test-02-03: / organization_repositories_url value = "

VALUE=$(curl -s "$BASE_URL" |jq -r '.organization_repositories_url')

if [[ "$VALUE" == "https://api.github.com/orgs/{org}/repos{?type,page,per_page,sort}" ]]; then
    pass
else
    fail "$VALUE" "https://api.github.com/orgs/{org}/repos{?type,page,per_page,sort}"
fi

describe "test-03-01: /orgs/Netflix key count = "

COUNT=$(curl -s "$BASE_URL/orgs/Netflix" |jq -r 'keys |.[]' |wc -l |awk '{print $1}')

if [[ $COUNT -eq 28 ]]; then
    pass
else
    fail "$COUNT" "28"
fi

describe "test-03-02: /orgs/Netflix avatar_url = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix" |jq -r '.avatar_url')

if [[ "$VALUE" == "https://avatars3.githubusercontent.com/u/913567?v=4" ]]; then
    pass
else
    fail "$VALUE" "https://avatars3.githubusercontent.com/u/913567?v=4"
fi

describe "test-03-03: /orgs/Netflix location = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix" |jq -r '.location')

if [[ "$VALUE" == "Los Gatos, California" ]]; then
    pass
else
    fail "$VALUE" "Los Gatos, California"
fi

describe "test-04-01: /orgs/Netflix/members object count = "

COUNT=$(curl -s "$BASE_URL/orgs/Netflix/members" |jq -r '. |length')

if [[ $COUNT -gt 6 ]] && [[ $COUNT -lt 10 ]]; then
    pass
else
    fail "$COUNT" "6..10"
fi

describe "test-04-02: /orgs/Netflix/members login first alpha case-insensitive = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix/members" |jq -r '.[] |.login' |tr '[:upper:]' '[:lower:]' |sort |head -1)

if [[ "$VALUE" == "chali" ]]; then
    pass
else
    fail "$VALUE" "chali"
fi

describe "test-04-03: /orgs/Netflix/members login first alpha case-sensitive = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix/members" |jq -r '.[] |.login' |sort |head -1)

if [[ "$VALUE" == "chali" ]]; then
    pass
else
    fail "$VALUE" "chali"
fi

describe "test-04-04: /orgs/Netflix/members login last alpha case-insensitive = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix/members" |jq -r '.[] |.login' |tr '[:upper:]' '[:lower:]' |sort |tail -1)

if [[ "$VALUE" == "rpalcolea" ]]; then
    pass
else
    fail "$VALUE" "rpalcolea"
fi

describe "test-04-05: /orgs/Netflix/members id first = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix/members" |jq -r '.[] |.id' |sort -n |head -1)

if [[ "$VALUE" == "217030" ]]; then
    pass
else
    fail "$VALUE" "217030"
fi

describe "test-04-06: /orgs/Netflix/members id last = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix/members" |jq -r '.[] |.id' |sort -n |tail -1)

if [[ "$VALUE" == "3108309" ]]; then
    pass
else
    fail "$VALUE" "3108309"
fi

describe "test-04-07: /users/chali/orgs proxy = "

VALUE=$(curl -s "$BASE_URL/users/chali/orgs" |jq -r '.[] |.login' |tr '\n' ':')

if [[ "$VALUE" == "Netflix:nebula-plugins:" ]]; then
    pass
else
    fail "$VALUE" "Netflix:nebula-plugins:"
fi

describe "test-04-08: /users/rpalcolea/orgs proxy = "

VALUE=$(curl -s "$BASE_URL/users/rpalcolea/orgs" |jq -r '.[] |.login' |tr '\n' ':')

if [[ "$VALUE" == "Netflix:nebula-plugins:" ]]; then
    pass
else
    fail "$VALUE" "Netflix:nebula-plugins:"
fi

describe "test-05-01: /orgs/Netflix/repos object count = "

COUNT=$(curl -s "$BASE_URL/orgs/Netflix/repos" |jq -r '. |length')

if [[ $COUNT -gt 127 ]] && [[ $COUNT -lt 177 ]]; then
    pass
else
    fail "$COUNT" "127..177"
fi

describe "test-05-02: /orgs/Netflix/repos full_name first alpha case-insensitive = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix/repos" |jq -r '.[] |.full_name' |tr '[:upper:]' '[:lower:]' |sort |head -1)

if [[ "$VALUE" == "netflix/aegisthus" ]]; then
    pass
else
    fail "$VALUE" "netflix/aegisthus"
fi

describe "test-05-03: /orgs/Netflix/members full_name first alpha case-sensitive = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix/repos" |jq -r '.[] |.full_name' |sort |head -1)

if [[ "$VALUE" == "Netflix/AWSObjectMapper" ]]; then
    pass
else
    fail "$VALUE" "Netflix/AWSObjectMapper"
fi

describe "test-05-04: /orgs/Netflix/members login last alpha case-insensitive = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix/repos" |jq -r '.[] |.full_name' |tr '[:upper:]' '[:lower:]' |sort |tail -1)

if [[ "$VALUE" == "netflix/zuul" ]]; then
    pass
else
    fail "$VALUE" "netflix/zuul"
fi

describe "test-05-05: /orgs/Netflix/repos id first = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix/repos" |jq -r '.[] |.id' |sort -n |head -1)

if [[ "$VALUE" == "2044029" ]]; then
    pass
else
    fail "$VALUE" "2044029"
fi

describe "test-05-06: /orgs/Netflix/repos id last = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix/repos" |jq -r '.[] |.id' |sort -n |tail -1)

if [[ "$VALUE" == "184627693" ]]; then
    pass
else
    fail "$VALUE" "184627693"
fi

describe "test-05-07: /orgs/Netflix/repos languages unique = "

VALUE=$(curl -s "$BASE_URL/orgs/Netflix/repos" |jq -r '.[] |.language' |sort -u |tr '\n' ':')

if [[ "$VALUE" == "C:C#:C++:Clojure:D:Dockerfile:Go:Groovy:HTML:Java:JavaScript:Python:Ruby:Scala:Shell:null:" ]]; then
    pass
else
    fail "$VALUE" "C:C#:C++:Clojure:D:Dockerfile:Go:Groovy:HTML:Java:JavaScript:Python:Ruby:Scala:Shell:null:"
fi

describe "test-06-01: /view/top/5/forks = "

VALUE=$(curl -s "$BASE_URL/view/top/5/forks" |tr -d '\n' |sed -e 's/ //g')

if [[ "$VALUE" == '[["Netflix/Hystrix",3518],["Netflix/eureka",2077],["Netflix/zuul",1445],["Netflix/SimianArmy",1039],["Netflix/ribbon",711]]' ]]; then
    pass
else
    fail "$VALUE" '[["Netflix/Hystrix",3518],["Netflix/eureka",2077],["Netflix/zuul",1445],["Netflix/SimianArmy",1039],["Netflix/ribbon",711]]'
fi

describe "test-06-02: /view/top/10/forks = "

VALUE=$(curl -s "$BASE_URL/view/top/10/forks" |tr -d '\n' |sed -e 's/ //g')

if [[ "$VALUE" == '[["Netflix/Hystrix",3518],["Netflix/eureka",2077],["Netflix/zuul",1445],["Netflix/SimianArmy",1039],["Netflix/ribbon",711],["Netflix/security_monkey",673],["Netflix/conductor",538],["Netflix/Cloud-Prize",488],["Netflix/chaosmonkey",455],["Netflix/asgard",437]]' ]]; then
    pass
else
    fail "$VALUE" '[["Netflix/Hystrix",3518],["Netflix/eureka",2077],["Netflix/zuul",1445],["Netflix/SimianArmy",1039],["Netflix/ribbon",711],["Netflix/security_monkey",673],["Netflix/conductor",538],["Netflix/Cloud-Prize",488],["Netflix/chaosmonkey",455],["Netflix/asgard",437]]'
fi

describe "test-06-03: /view/top/5/last_updated = "

VALUE=$(curl -s "$BASE_URL/view/top/5/last_updated" |tr -d '\n' |sed -e 's/ //g')

if [[ "$VALUE" == '[["Netflix/dynomite","2019-05-12T19:18:57Z"],["Netflix/SimianArmy","2019-05-12T19:13:22Z"],["Netflix/Hystrix","2019-05-12T18:50:04Z"],["Netflix/vmaf","2019-05-12T18:17:29Z"],["Netflix/fast_jsonapi","2019-05-12T17:28:08Z"]]' ]]; then
    pass
else
    fail "$VALUE" '[["Netflix/dynomite","2019-05-12T19:18:57Z"],["Netflix/SimianArmy","2019-05-12T19:13:22Z"],["Netflix/Hystrix","2019-05-12T18:50:04Z"],["Netflix/vmaf","2019-05-12T18:17:29Z"],["Netflix/fast_jsonapi","2019-05-12T17:28:08Z"]]'
fi

describe "test-06-04: /view/top/10/last_updated = "

VALUE=$(curl -s "$BASE_URL/view/top/10/last_updated" |tr -d '\n' |sed -e 's/ //g')

if [[ "$VALUE" == '[["Netflix/dynomite","2019-05-12T19:18:57Z"],["Netflix/SimianArmy","2019-05-12T19:13:22Z"],["Netflix/Hystrix","2019-05-12T18:50:04Z"],["Netflix/vmaf","2019-05-12T18:17:29Z"],["Netflix/fast_jsonapi","2019-05-12T17:28:08Z"],["Netflix/eureka","2019-05-12T16:59:05Z"],["Netflix/flamescope","2019-05-12T16:49:11Z"],["Netflix/security_monkey","2019-05-12T16:43:38Z"],["Netflix/chaosmonkey","2019-05-12T16:26:56Z"],["Netflix/lemur","2019-05-12T16:12:16Z"]]' ]]; then
    pass
else
    fail "$VALUE" '[["Netflix/dynomite","2019-05-12T19:18:57Z"],["Netflix/SimianArmy","2019-05-12T19:13:22Z"],["Netflix/Hystrix","2019-05-12T18:50:04Z"],["Netflix/vmaf","2019-05-12T18:17:29Z"],["Netflix/fast_jsonapi","2019-05-12T17:28:08Z"],["Netflix/eureka","2019-05-12T16:59:05Z"],["Netflix/flamescope","2019-05-12T16:49:11Z"],["Netflix/security_monkey","2019-05-12T16:43:38Z"],["Netflix/chaosmonkey","2019-05-12T16:26:56Z"],["Netflix/lemur","2019-05-12T16:12:16Z"]]'
fi

describe "test-06-05: /view/top/5/open_issues = "

VALUE=$(curl -s "$BASE_URL/view/top/5/open_issues" |tr -d '\n' |sed -e 's/ //g')

if [[ "$VALUE" == '[["Netflix/Hystrix",335],["Netflix/ribbon",162],["Netflix/astyanax",161],["Netflix/zuul",139],["Netflix/eureka",116]]' ]]; then
    pass
else
    fail "$VALUE" '[["Netflix/Hystrix",335],["Netflix/ribbon",162],["Netflix/astyanax",161],["Netflix/zuul",139],["Netflix/eureka",116]]'
fi

describe "test-06-06: /view/top/10/open_issues = "

VALUE=$(curl -s "$BASE_URL/view/top/10/open_issues" |tr -d '\n' |sed -e 's/ //g')

if [[ "$VALUE" == '[["Netflix/Hystrix",335],["Netflix/ribbon",162],["Netflix/astyanax",161],["Netflix/zuul",139],["Netflix/eureka",116],["Netflix/asgard",105],["Netflix/conductor",104],["Netflix/archaius",87],["Netflix/security_monkey",84],["Netflix/fast_jsonapi",75]]' ]]; then
    pass
else
    fail "$VALUE" '[["Netflix/Hystrix",335],["Netflix/ribbon",162],["Netflix/astyanax",161],["Netflix/zuul",139],["Netflix/eureka",116],["Netflix/asgard",105],["Netflix/conductor",104],["Netflix/archaius",87],["Netflix/security_monkey",84],["Netflix/fast_jsonapi",75]]'
fi

describe "test-06-07: /view/top/5/stars = "

VALUE=$(curl -s "$BASE_URL/view/top/5/stars" |tr -d '\n' |sed -e 's/ //g')

if [[ "$VALUE" == '[["Netflix/Hystrix",17256],["Netflix/falcor",9318],["Netflix/eureka",7685],["Netflix/pollyjs",7630],["Netflix/zuul",7437]]' ]]; then
    pass
else
    fail "$VALUE" '[["Netflix/Hystrix",17256],["Netflix/falcor",9318],["Netflix/eureka",7685],["Netflix/pollyjs",7630],["Netflix/zuul",7437]]'
fi

describe "test-06-08: /view/top/10/stars = "

VALUE=$(curl -s "$BASE_URL/view/top/10/stars" |tr -d '\n' |sed -e 's/ //g')

if [[ "$VALUE" == '[["Netflix/Hystrix",17256],["Netflix/falcor",9318],["Netflix/eureka",7685],["Netflix/pollyjs",7630],["Netflix/zuul",7437],["Netflix/SimianArmy",7137],["Netflix/chaosmonkey",6371],["Netflix/fast_jsonapi",4262],["Netflix/security_monkey",3531],["Netflix/vector",3121]]' ]]; then
    pass
else
    fail "$VALUE" '[["Netflix/Hystrix",17256],["Netflix/falcor",9318],["Netflix/eureka",7685],["Netflix/pollyjs",7630],["Netflix/zuul",7437],["Netflix/SimianArmy",7137],["Netflix/chaosmonkey",6371],["Netflix/fast_jsonapi",4262],["Netflix/security_monkey",3531],["Netflix/vector",3121]]'
fi

report
