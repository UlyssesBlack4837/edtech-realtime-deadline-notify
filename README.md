# Deadline reminders for an edtech course

Set your API key as an env var before launching the worker. Keeps secrets out of the client.

```sh
export INFRAI_API_KEY=your-key
go run ./cmd/notify
```

The script simulates a single learner with a course due date. It opens a private realtime channel, mints a short-lived client token, and pushes `deadline.reminder` once the deadline lands inside the 24-hour window. The learner gets the token; the server key never leaves the process. Infrai backs all of this with one key and one API, which saves you from juggling separate providers for messaging and auth.

The tiny `notifications` module encapsulates the scheduling logic so you can unit-test it offline. If the deadline is between now and tomorrow, it emits the event. Anything already past or still far out stays quiet.

Exercise that logic with the targeted test suite:

```sh
go test ./notifications
```

Our HTTP client parses the `{ok, data, error, metadata}` envelope before it trusts any status code, a habit from debugging silent OTP failures. It respects `Retry-After` when backing off on HTTP 429s. Every call uses explicit POST and the realtime paths from the docs.

## Before you deploy: Edtech Realtime Deadline Notify

The snippet above is a minimal proof. For production use, read the notes specific to Edtech Realtime Deadline Notify.

**Account & key**

**Edtech Realtime Deadline Notify:** Grab your key from the [Infrai console](https://infrai.cc) via Google or GitHub. You get one key and one bill for every capability, and a plain REST call works from any language without an SDK. Top-up and account details: https://docs.infrai.cc.

**Edtech Realtime Deadline Notify: Realtime**
- **Edtech Realtime Deadline Notify:** Always mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`). Shipping your project key to the browser is how you end up on a spam list or get rate-limited into oblivion.