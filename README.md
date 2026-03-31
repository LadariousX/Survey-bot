# Discord chatbot for solving fast food surveys.
## How it works
### Whataburger
Send QR or link from QR on recept. solver solves the survey and sends the cupon for free burger with a purchase of fries and drink to the email given.
### Dairy Queen
Send QR or link from QR on recept. solver returns a confirmation image continuing free dilly bar cupon.
### Canes
Use CLI do deliver code to bot. solver enters user in the Free canes for a year drawing. 

## External dependencies
Create a .env file and fill it with these values:
Ensure you are using the bot token from the "Bot" tab and not the "Client Secret" from the "OAuth2" tab.

requirements for each solver:

| Solver      | requirement                        |
|-------------|------------------------------------|
| All         | - Chome or chromium<br/>- Discord  |
| Whataburger | [Capsolver api key](capsolver.com) |
| Dairy Queen | none                               |
| Canes       | idk might be broke                 |
```
BotToken=
DefaultEmail=
CapSolverKey=
```
Run with `go run main.go`

For QR reader capabilities set up the venv and install python requirements:
```aiignore
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```
Docker build command: `docker build -t ladariousx/survey-bot:v2.01 .;` Unfortunetly the being image is super bloated beucause of base
image and opencv. BEWARE the .env is copied to the image and therefor container.