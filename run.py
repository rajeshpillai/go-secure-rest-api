"""Run every chapter's server and the exact commands shown on screen; record what they printed.

    python3 code/run.py            # all chapters
    python3 code/run.py ch05 ch06  # some

Each demo writes code/out/<chapter>/<name>.txt: the command line(s) as typed ("$ curl ..."),
then the real output. curl's -i headers are kept except Date (it changes every run).
scenes.mjs reads these files; nothing on screen is typed by hand.
"""
import json
import os
import re
import shlex
import signal
import socket
import subprocess
import sys
import time
from pathlib import Path

HERE = Path(__file__).resolve().parent
OUT = HERE / "out"
BIN = HERE / ".bin"
SECRET = "k7Hq2vN9xR4mW8zT1pL6sD3fG5jB0cYe"  # demo-only JWT key, passed via the environment like a real one
ENV = {**os.environ, "JWT_SECRET": SECRET}


def build(ch):
    BIN.mkdir(exist_ok=True)
    subprocess.run(["go", "build", "-o", str(BIN / ch), f"./{ch}"], cwd=HERE, check=True)


class Server:
    def __init__(self, ch, env=None, port=8080, args=()):
        self.ch, self.port = ch, port
        self.log = OUT / ch / "server.log"
        self.log.parent.mkdir(parents=True, exist_ok=True)
        self.f = open(self.log, "w")
        self.p = subprocess.Popen([str(BIN / ch), *args], cwd=HERE / ch, env=env or ENV,
                                  stdout=self.f, stderr=subprocess.STDOUT)
        for _ in range(200):
            try:
                socket.create_connection(("127.0.0.1", port), 0.1).close()
                return
            except OSError:
                if self.p.poll() is not None:
                    raise SystemExit(f"{ch} exited: {self.log.read_text()}")
                time.sleep(0.05)
        raise SystemExit(f"{ch} did not start")

    def stop(self):
        self.p.send_signal(signal.SIGTERM)
        self.p.wait(5)
        self.f.close()
        return self.log.read_text()


def clean(s):
    s = re.sub(r"^date: .*\r?\n", "", s, flags=re.M | re.I)
    s = re.sub(r"^\d{4}/\d\d/\d\d \d\d:\d\d:\d\d ", "", s, flags=re.M)
    s = s.replace(str(HERE) + "/", "./").replace(str(HERE), ".")
    return s.replace("\r\n", "\n").rstrip("\n")


def sh(cmd, env=None):
    """Run one shell line, return its stdout+stderr (cleaned)."""
    r = subprocess.run(["bash", "-c", cmd], cwd=HERE, env=env or ENV, capture_output=True, text=True)
    return clean(r.stdout + r.stderr)


def save(ch, name, pairs):
    """pairs: [(shown command, output)] → out/ch/name.txt"""
    d = OUT / ch
    d.mkdir(parents=True, exist_ok=True)
    txt = "\n".join(f"$ {c}\n{o}" if o else f"$ {c}" for c, o in pairs)
    (d / f"{name}.txt").write_text(txt + "\n")
    return txt


def demo(ch, name, *cmds, env=None, cwd=None):
    """Run the commands in ONE bash session (variables like TOKEN carry over)."""
    script = "".join(f"echo __CMD_{i}__\n{c}\n" for i, c in enumerate(cmds))
    raw = subprocess.run(["bash", "-c", script], cwd=cwd or HERE, env=env or ENV, capture_output=True, text=True)
    parts = re.split(r"__CMD_\d+__\n", raw.stdout + raw.stderr)[1:]
    pairs = [(c, clean(o)) for c, o in zip(cmds, parts)]
    print(f"--- {ch}/{name}\n" + save(ch, name, pairs))
    return [o for _, o in pairs]


def body(out):
    """The JSON body of a `curl -si` output."""
    return json.loads(out.split("\n\n", 1)[1]) if "\n\n" in out else json.loads(out)


# ---------------------------------------------------------------- chapters
def ch03():
    build("ch03")
    s = Server("ch03")
    demo("ch03", "hello", "curl -s localhost:8080/")
    s.stop()


def ch04():
    build("ch04")
    s = Server("ch04")
    demo("ch04", "list", "curl -s localhost:8080/books")
    demo("ch04", "one", "curl -s localhost:8080/books/2")
    demo("ch04", "missing", "curl -si localhost:8080/books/99")
    demo("ch04", "wrong_method", "curl -si -X DELETE localhost:8080/books/2")
    s.stop()


def ch05():
    build("ch05")
    s = Server("ch05")
    demo("ch05", "list", "curl -s localhost:8080/books")
    demo("ch05", "create", """curl -si -X POST localhost:8080/books -d '{"title":"Go in Action","author":"William Kennedy"}'""")
    demo("ch05", "no_title", """curl -si -X POST localhost:8080/books -d '{"author":"Nobody"}'""")
    demo("ch05", "unknown_field", """curl -si -X POST localhost:8080/books -d '{"title":"Hack","id":1}'""")
    demo("ch05", "bad_json", """curl -si -X POST localhost:8080/books -d '{"title":'""")
    demo("ch05", "bad_id", "curl -si localhost:8080/books/abc")
    s.stop()
    demo("ch05", "race", "go test -race ./ch05 2>&1 | grep -m1 -A3 'DATA RACE'; go test -race ./ch05 2>&1 | tail -1")


def ch06():
    build("ch06")
    s = Server("ch06")
    demo("ch06", "update", """curl -si -X PUT localhost:8080/books/2 -d '{"title":"Clean Code (2nd ed.)","author":"Robert C. Martin"}'""")
    demo("ch06", "delete", "curl -si -X DELETE localhost:8080/books/3")
    demo("ch06", "list", "curl -s localhost:8080/books")
    s.stop()
    demo("ch06", "race", "go test -race -count=1 ./ch06")
    # the cold open: the learner's API, no locks on the door
    s = Server("ch06")
    demo("ch06", "cold_list", "curl -s localhost:8080/books")
    demo("ch06", "cold_wipe", "for id in 1 2 3; do curl -s -X DELETE localhost:8080/books/$id; done",
         "curl -s localhost:8080/books")
    s.stop()


def ch07():
    build("ch07")
    s = Server("ch07")
    demo("ch07", "get", "curl -s localhost:8080/books/1")
    demo("ch07", "missing", "curl -s localhost:8080/books/42")
    demo("ch07", "oops", "curl -si localhost:8080/oops")
    demo("ch07", "after_oops", "curl -s localhost:8080/books/2")
    # a slow client: opens a connection, sends half a request line, then waits
    t0 = time.time()
    c = socket.create_connection(("127.0.0.1", 8080))
    c.sendall(b"GET /books HTTP/1.1\r\n")
    c.settimeout(30)
    data = c.recv(4096)
    held = time.time() - t0
    c.close()
    (OUT / "ch07").mkdir(parents=True, exist_ok=True)
    (OUT / "ch07" / "slow_client.json").write_text(json.dumps({"seconds": round(held, 1), "reply": data.decode(errors="replace")}))
    print("--- ch07/slow_client", round(held, 1), "s", data[:60])
    log = s.stop()
    (OUT / "ch07" / "log.txt").write_text(re.sub(r"^\S+ \S+ ", "", log, flags=re.M))


BELLA = """'{"email":"bella@example.com","password":"bella-loves-old-books"}'"""
MALLORY = """'{"email":"mallory@example.com","password":"mallory-wants-your-books"}'"""


def fresh(ch):
    (HERE / ch / "users.json").unlink(missing_ok=True)


def ch09plain():
    build("ch09plain")
    fresh("ch09plain")
    s = Server("ch09plain")
    demo("ch09plain", "register",
         """curl -s -X POST localhost:8080/register -d '{"email":"bella@example.com","password":"bella-loves-old-books"}'""",
         """curl -s -X POST localhost:8080/register -d '{"email":"sam@example.com","password":"correct horse battery staple"}'""")
    demo("ch09plain", "leak", "cat users.json", cwd=HERE / "ch09plain")
    s.stop()


def ch09():
    build("ch09")
    fresh("ch09")
    s = Server("ch09")
    demo("ch09", "register",
         """curl -si -X POST localhost:8080/register -d '{"email":"bella@example.com","password":"bella-loves-old-books"}'""")
    demo("ch09", "short", """curl -s -X POST localhost:8080/register -d '{"email":"ana@example.com","password":"hunter2"}'""")
    demo("ch09", "same_password",
         """curl -s -X POST localhost:8080/register -d '{"email":"sam@example.com","password":"correct horse battery staple"}'""",
         """curl -s -X POST localhost:8080/register -d '{"email":"kim@example.com","password":"correct horse battery staple"}'""")
    demo("ch09", "users", "cat users.json", cwd=HERE / "ch09")
    demo("ch09", "login_ok", """curl -si -X POST localhost:8080/login -d '{"email":"bella@example.com","password":"bella-loves-old-books"}'""")
    demo("ch09", "login_bad", """curl -si -X POST localhost:8080/login -d '{"email":"bella@example.com","password":"bella-loves-new-books"}'""")
    demo("ch09", "login_nobody", """curl -s -X POST localhost:8080/login -d '{"email":"nobody@example.com","password":"bella-loves-old-books"}'""")
    # timing: an unknown email must cost the same as a wrong password
    tm_cmds = ("""curl -s -o /dev/null -w '%{time_total}s\\n' -X POST localhost:8080/login -d '{"email":"bella@example.com","password":"wrong-password-here"}'""",
         """curl -s -o /dev/null -w '%{time_total}s\\n' -X POST localhost:8080/login -d '{"email":"nobody@example.com","password":"wrong-password-here"}'""")
    demo("ch09", "timing", *tm_cmds)
    s.stop()
    # the same login without the dummy hash, to show what it prevents (a scratch copy, never shown as code)
    tmp = HERE / "nodummy"
    subprocess.run(["rm", "-rf", str(tmp)])
    subprocess.run(["cp", "-r", str(HERE / "ch09"), str(tmp)], check=True)
    src = (tmp / "auth.go").read_text()
    cut = "\tif !found {\n\t\tcheckPassword(in.Password, dummyHash)\n\t}\n"
    assert cut in src
    (tmp / "auth.go").write_text(src.replace(cut, ""))
    subprocess.run(["go", "build", "-o", str(BIN / "nodummy"), "./nodummy"], cwd=HERE, check=True)
    (tmp / "users.json").unlink(missing_ok=True)
    s = Server("nodummy")
    sh("curl -s -X POST localhost:8080/register -d " + BELLA)
    demo("ch09", "timing_nodummy", *tm_cmds)
    s.stop()
    subprocess.run(["rm", "-rf", str(tmp), str(OUT / "nodummy")])
    demo("ch09", "bench", "go test -bench . -benchtime 50x ./ch09 | grep Benchmark")


def ch11():
    build("ch11")
    fresh("ch11")
    demo("ch11", "no_secret", "JWT_SECRET= ./.bin/ch11")
    s = Server("ch11")
    demo("ch11", "setup",
         """curl -s -X POST localhost:8080/register -d '{"email":"bella@example.com","password":"bella-loves-old-books"}'""")
    demo("ch11", "login", """curl -s -X POST localhost:8080/login -d '{"email":"bella@example.com","password":"bella-loves-old-books"}'""")
    login = """TOKEN=$(curl -s -X POST localhost:8080/login -d '{"email":"bella@example.com","password":"bella-loves-old-books"}' | jq -r .access_token)"""
    demo("ch11", "parts", login, "echo $TOKEN | cut -d. -f1 | basenc -d --base64url",
         "echo $TOKEN | cut -d. -f2 | basenc -d --base64url")
    demo("ch11", "no_token", """curl -si -X POST localhost:8080/books -d '{"title":"Go in Action"}'""")
    demo("ch11", "with_token", login,
         """curl -si -X POST localhost:8080/books -H "Authorization: Bearer $TOKEN" -d '{"title":"Go in Action"}'""")
    demo("ch11", "me", login, 'curl -s localhost:8080/me -H "Authorization: Bearer $TOKEN"')
    # tamper: same header and signature, payload says user 2
    demo("ch11", "tamper", login,
         """FAKE=$(echo -n '{"iss":"bookshelf","sub":"2","exp":1999999999,"iat":1700000000}' | basenc -w0 --base64url | tr -d '=')""",
         """BAD=$(echo $TOKEN | cut -d. -f1).$FAKE.$(echo $TOKEN | cut -d. -f3)""",
         'curl -si localhost:8080/me -H "Authorization: Bearer $BAD"')
    demo("ch11", "none", "NONE=$(go run ./tools/mint none)", "echo $NONE | cut -d. -f1 | basenc -d --base64url 2>/dev/null",
         'curl -s localhost:8080/me -H "Authorization: Bearer $NONE"')
    demo("ch11", "expired", "OLD=$(go run ./tools/mint expired)",
         'curl -s localhost:8080/me -H "Authorization: Bearer $OLD"')
    log = s.stop()
    (OUT / "ch11" / "log.txt").write_text(re.sub(r"^\S+ \S+ ", "", log, flags=re.M))
    print(log)




def login_cmd(var, creds, url="localhost:8080", extra=""):
    return f"{var}=$(curl -s{extra} -X POST {url}/login -d {creds} | jq -r .access_token)"


def ch12():
    build("ch12")
    fresh("ch12")
    s = Server("ch12")
    sh(f"curl -s -X POST localhost:8080/register -d {BELLA}")
    demo("ch12", "login", f"curl -s -X POST localhost:8080/login -d {BELLA}")
    demo("ch12", "rotate",
         f"R1=$(curl -s -X POST localhost:8080/login -d {BELLA} | jq -r .refresh_token)",
         """R2=$(curl -s -X POST localhost:8080/refresh -d "{\\"refresh_token\\":\\"$R1\\"}" | jq -r .refresh_token); echo $R2""",
         """curl -si -X POST localhost:8080/refresh -d "{\\"refresh_token\\":\\"$R1\\"}\"""",
         """curl -s -X POST localhost:8080/refresh -d "{\\"refresh_token\\":\\"$R2\\"}\"""")
    demo("ch12", "logout",
         f"""LOGIN=$(curl -s -X POST localhost:8080/login -d {BELLA})""",
         "TOKEN=$(echo $LOGIN | jq -r .access_token); R=$(echo $LOGIN | jq -r .refresh_token)",
         """curl -si -X POST localhost:8080/logout -d "{\\"refresh_token\\":\\"$R\\"}\"""",
         """curl -s -X POST localhost:8080/refresh -d "{\\"refresh_token\\":\\"$R\\"}\"""",
         'curl -s localhost:8080/me -H "Authorization: Bearer $TOKEN"')
    s.stop()


def ch13():
    build("ch13")
    fresh("ch13")
    s = Server("ch13")
    sh(f"curl -s -X POST localhost:8080/register -d {BELLA}")
    tok = login_cmd("TOKEN", BELLA)
    demo("ch13", "create", tok,
         """curl -s -X POST localhost:8080/keys -H "Authorization: Bearer $TOKEN" -d '{"name":"backup script","scopes":["books:read"]}' | jq .""")
    demo("ch13", "use", tok,
         """KEY=$(curl -s -X POST localhost:8080/keys -H "Authorization: Bearer $TOKEN" -d '{"name":"reader","scopes":["books:read"]}' | jq -r .key)""",
         'curl -s localhost:8080/books -H "X-API-Key: $KEY"',
         """curl -si -X POST localhost:8080/books -H "X-API-Key: $KEY" -d '{"title":"Hack"}'""",
         """curl -s -X POST localhost:8080/keys -H "X-API-Key: $KEY" -d '{"name":"more","scopes":["books:write"]}'""",
         'curl -s localhost:8080/me -H "X-API-Key: $KEY"')
    demo("ch13", "list", tok, 'curl -s localhost:8080/keys -H "Authorization: Bearer $TOKEN" | jq -c ".[]"')
    demo("ch13", "revoke", tok,
         """KEY=$(curl -s -X POST localhost:8080/keys -H "Authorization: Bearer $TOKEN" -d '{"name":"old laptop","scopes":["books:read"]}' | jq -r .key)""",
         """ID=$(curl -s localhost:8080/keys -H "Authorization: Bearer $TOKEN" | jq '.[-1].id')""",
         'curl -si -X DELETE localhost:8080/keys/$ID -H "Authorization: Bearer $TOKEN"',
         'curl -s localhost:8080/books -H "X-API-Key: $KEY"')
    s.stop()


def bola(ch):
    build(ch)
    fresh(ch)
    s = Server(ch)
    sh(f"curl -s -X POST localhost:8080/register -d {BELLA}")
    sh(f"curl -s -X POST localhost:8080/register -d {MALLORY}")
    demo(ch, "bella_adds", login_cmd("BELLA", BELLA),
         """curl -s -X POST localhost:8080/books -H "Authorization: Bearer $BELLA" -d '{"title":"My diary","author":"Bella"}'""")
    demo(ch, "mallory_list", login_cmd("MALLORY", MALLORY),
         'curl -s localhost:8080/books -H "Authorization: Bearer $MALLORY"')
    demo(ch, "mallory_reads", login_cmd("MALLORY", MALLORY),
         'curl -si localhost:8080/books/1 -H "Authorization: Bearer $MALLORY"')
    demo(ch, "mallory_deletes", login_cmd("MALLORY", MALLORY),
         'curl -si -X DELETE localhost:8080/books/1 -H "Authorization: Bearer $MALLORY"')
    demo(ch, "bella_list", login_cmd("BELLA", BELLA),
         'curl -s localhost:8080/books -H "Authorization: Bearer $BELLA"')
    s.stop()


def ch14bug():
    bola("ch14bug")


def ch14():
    bola("ch14")


def ch15():
    build("ch15")
    fresh("ch15")
    s = Server("ch15")
    sh(f"curl -s -X POST localhost:8080/register -d {BELLA}")
    demo("ch15", "guessing",
         """for i in 1 2 3 4 5 6 7; do curl -s -o /dev/null -w "try $i: %{http_code}\\n" -X POST localhost:8080/login -d '{"email":"bella@example.com","password":"is-it-password-'$i'"}'; done""")
    demo("ch15", "retry_after",
         f"curl -si -X POST localhost:8080/login -d {BELLA}")
    s.stop()


def ch16():
    build("ch16")
    fresh("ch16")
    d = HERE / "ch16"
    for f in ("cert.pem", "key.pem"):
        (d / f).unlink(missing_ok=True)
    demo("ch16", "cert", "go run $(go env GOROOT)/src/crypto/tls/generate_cert.go --host localhost --ca", cwd=d)
    env = {**ENV, "CORS_ORIGIN": "https://bookshelf.example"}
    s = Server("ch16", env=env, port=8443)
    demo("ch16", "plain_http", "curl -s http://localhost:8443/books", cwd=d)
    demo("ch16", "untrusted", "curl -sS https://localhost:8443/books", cwd=d)
    demo("ch16", "https", "curl -s --cacert cert.pem https://localhost:8443/books", cwd=d)
    demo("ch16", "cors_ok",
         """curl -si -X OPTIONS https://localhost:8443/books --cacert cert.pem -H "Origin: https://bookshelf.example" -H "Access-Control-Request-Method: POST\"""", cwd=d)
    demo("ch16", "cors_other",
         """curl -si -X OPTIONS https://localhost:8443/books --cacert cert.pem -H "Origin: https://evil.example" -H "Access-Control-Request-Method: POST\"""", cwd=d)
    demo("ch16", "final_wipe",
         """for id in 1 2 3; do curl -s -o /dev/null -w "%{http_code}\\n" --cacert cert.pem -X DELETE https://localhost:8443/books/$id; done""", cwd=d)
    s.stop()
    demo("ch16", "no_secret", "JWT_SECRET= ../.bin/ch16", cwd=d)
    demo("ch16", "gitignore", "cat .gitignore", cwd=HERE)


def ch17():
    demo("ch17", "test", "go test -race -count=1 -v ./ch17 2>&1 | grep -E '^(---|    ---|ok|PASS|FAIL)'")
    # bring Bella's bug back in a scratch copy: the tests must catch it
    tmp = HERE / ".bugcheck"
    subprocess.run(["rm", "-rf", str(tmp)])
    subprocess.run(["cp", "-r", str(HERE / "ch17"), str(tmp)], check=True)
    src = (tmp / "books.go").read_text()
    cut = """	if b.OwnerID != current(r).UserID {
		writeError(w, http.StatusForbidden, "not your book")
		return Book{}, false
	}
"""
    assert cut in src
    (tmp / "books.go").write_text(src.replace(cut, ""))
    out = sh("go test -count=1 ./.bugcheck 2>&1 | grep -E '^(---|    ---|        |ok|FAIL)'")
    out = out.replace("./.bugcheck", "./ch17").replace("bookshelf/.bugcheck", "bookshelf/ch17")
    print("--- ch17/bug_caught\n" + save("ch17", "bug_caught", [("go test ./ch17     # with the owner check removed", out)]))
    subprocess.run(["rm", "-rf", str(tmp)])


CHAPTERS = {f.__name__: f for f in [ch03, ch04, ch05, ch06, ch07, ch09plain, ch09, ch11, ch12, ch13, ch14bug, ch14, ch15, ch16, ch17]}

if __name__ == "__main__":
    for name in sys.argv[1:] or CHAPTERS:
        CHAPTERS[name]()
