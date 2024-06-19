# OAuth Discord Trade
Generates an access token through discord's API code exchange.
To get the OAuth2 code, you need to:
- Go to https://discord.com/developers/applications
- Create new application
- Go to OAuth2 tab
- Add 'http://localhost:{port}' to redirects
- Generate link with necessary scopes for your use case
- Start the application and go to the generated link
