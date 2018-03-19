# Jennie
## Email and messenger Assistant (under work)

Under progress email and messenger Assistant

### Phase 1: Jennie supports auto responding to Linkedin Messages

Since, I do not check Linkedin frequently a lot of messages go un-responded to. I was looking for a way to set an auto-reply whereby I can redirect them to send me an email instead with more details included. I checked Linkedin APIs and features but none seem to meet the requirements. Luckily, linkedin messages are by default sent to my Gmail, so this responds to Linkedin messages using Gmail apis

### Setup
1. Clone this repo
2. Create a google cloud project using Google Developer Console
3. In the project settings enable Google API using following steps

  * On the Add credentials to your project page, click the Cancel button.
  * At the top of the page, select the OAuth consent screen tab. Select an Email address, enter a Product name if not already set, and click the Save button.
  * Select the Credentials tab, click the Create credentials button and select OAuth client ID.
  * Select the application type Other, enter the name "Gmail API Quickstart", and click the Create button.
  * Click OK to dismiss the resulting dialog.
  * Click the file_download (Download JSON) button to the right of the client ID.
  * Move this file to your working directory and rename it client_secret.json and place it in the project directory
4. Rename `conf/conf.sample.json` to `conf/conf.json` and set your own 16 bit private key.
5. Run `go build`
6. Run `./jennie`
7. By default the server starts on port 8000
8. Open browser and navigate to `http://<server-ip>:8000/authorize`
9. Authorize with your email id and accept the permissions
10. That's it. Now it will start pulling messages from gmail which are being sent by Linkedin Messaging and respond to ones not responded to. Let the server run.

### TODO:
* Create a hosting and let it work for all users
* Move this assistant to mobile

### Next steps
* Scanning mails and getting a list of mails that needs responding to
* Extend this to auto respond to FB messenger and Hangout messages
* Add intelligence to auto replies
