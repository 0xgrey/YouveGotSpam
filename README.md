# YouveGotSpam
Multitool for auditing email security settings, sending spoofed emails as proof of concept, and automating mail send via Mailgun API.

Designed with a minimal reliance on third-party dependencies.

![YouveGotSpam Logo](/media/logo.png)

## TODO
 - [ ] Support for Mailgun API to send emails.
 - [ ] Function to send spoofed emails via authenticated relays, instead of just direct-send.
 - [ ] Add logging functionality via TXT, JSON, and CSV.
 - [x] Support for checking spoofable status against multiple domains via "investigate" command.
 - [x] Check spoofability status with domains pulled from MDI.
 - [x] Functionality to send spoofed emails when DMARC/SPF records insufficiently enforced.