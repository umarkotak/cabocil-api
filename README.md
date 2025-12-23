# Ca Bocil Backend Service

## MISC

- font awesome list: https://fontawesome.com/search?o=r&m=free
- channel id finder: https://commentpicker.com/youtube-channel-id.php

## PRE REQUISITES

- brew install caesiumclt
- brew install ghostscript

## Notes

if you run your golang app on wsl, you might need to get the wsl ip
`ip addr show eth0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}'`
source: https://github.com/postmanlabs/postman-app-support/issues/11204#issuecomment-1605502449

cloudflared access tcp --hostname pg.cabocil.com --url localhost:54322
