# Survey Bot

API that automates fast food customer survey forms to claim rewards.

## Supported surveys

| Survey | URL pattern | Reward |
|---|---|---|
| Whataburger | `whataburger.com` | Coupon emailed to provided address |
| Dairy Queen | `mydqexperience.com` | Verification code returned in response |
While there are chunks of legacy code for different businesses that may work,
they've been disabled until I can find the time to test them.
## Prerequisites

- Go 1.21+
- Chrome or Chromium installed
- capsolverapi key (needed for Whataburger)

## Setup
Create a `.env` file (or export the vars directly):

```
CapSolverKey=   # required for Whataburger only
```

## Start API
```sh
go run main.go
```

## API
**Example Whataburger request using URL from QR code on receipt**
```sh
curl -X POST http://localhost:3000/api/survey-bot \
  -H "Content-Type: application/json" \
  -d '{"url":"URLToSurveyFromQr", "email":"you@example.com"}'
```


### `POST /api/survey-bot`

**Request**
```json
{ "url":   "<survey url>",
  "email": "<your email>" }
```

**Response (200)**
```json
{ "message": "Survey completed successfully.", 
  "code":    "<verification code>" }
```
`"code"` is only present for DQ surveys.

