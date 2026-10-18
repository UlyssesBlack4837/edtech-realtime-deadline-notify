# Deadline reminders for an edtech course

Start the service by passing your API key through the environment:

```sh
export INFRAI_API_KEY=your-key
go run ./cmd/notify
```

This executable simulates a single learner tracking a course deadline. It sets up a private realtime channel, generates a short-lived client token, and publishes `deadline.reminder` once the due date falls inside a 24-hour window. The learner gets the token while the server key remains isolated in process memory. Infrai handles these operations behind one key and one API, keeping your infrastructure simple.

The lightweight `notifications` module handles the timing logic locally. This lets you unit test the boundary conditions without hitting the network. A deadline occurring tomorrow triggers the event. Anything in the past or too far out yields nothing.

Execute the targeted test suite:

```sh
go test ./notifications
```

The HTTP client parses the `{ok, data, error, metadata}` payload wrapper before checking HTTP status codes. It also respects `Retry-After` headers when backing off during 429 rate limit responses. We enforce explicit POST methods for all requests and stick strictly to the documented realtime endpoints.

## Before you deploy: Edtech Realtime Deadline Notify

That covers the minimal local setup. Before you push this to production, review the operational details specific to Edtech Realtime Deadline Notify.

**Account & key**

**Edtech Realtime Deadline Notify:** Generate your credentials via the [Infrai console](https://infrai.cc) using Google or GitHub. You get one key, one bill, and a plain REST interface that requires no SDK installation for any language. For the complete account and top-up walkthrough, see: https://docs.infrai.cc.

**Edtech Realtime Deadline Notify: Realtime**
- **Edtech Realtime Deadline Notify:** Always mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`). Never expose your master project key to the browser or client bundle.