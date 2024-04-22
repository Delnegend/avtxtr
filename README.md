# Avtxtr
Extract users' avatars from various social media

## Pre-requisite
- [Docker Engine](https://docs.docker.com/engine/install/)

## Deployment
- Clone this repository

- Change the environment variable
    | Variable | Description | Default |
    | --- | --- | --- |
    | `CHROME_ADDR` | [`chromedp`](https://github.com/chromedp/chromedp) address, only needed for Twitter | `http://chromedp:9222` |
    | `FA_COOKIE_A`, `FA_COOKIE_B` | cookies for FurAffinity | - |
    | `MAX_REQUEST_PER_TIME_UNIT` | maximum request per IP per time unit | `50` |
    | `CLEAR_LIST_EVERY_TIME_UNIT` | the above time unit | `1h` |
    | `REQUEST_TIMEOUT` | maximum time to wait for a request to complete | `12s` |

- `docker compose up -d`

## Supported socials
- DeviantArt
- FurAffinity
- Gravatar
- Instagram
- Threads
- Reddit
- Telegram
- Twitter/X
- YouTube

## API
- `GET localhost:8080/{social}/{username}?fallback={fallback_url}`
    - `social`: social media
    - `username`: username
    - `fallback_url`: the fallback avatar image
    - Response
        - `200`: the avatar image
        - `429`: too many requests
        - `500`: internal server error
    - Example
        - `GET localhost:8080/deviantart/username`
        - `GET localhost:8080/deviantart/username?fallback=https://example.com/avatar.png`

## FAQ
- How to get the cookies for FurAffinity?
    - Open the website
    - Login
    - Open the developer console
    - Go to the `Application` tab
    - Find the cookies
    - Copy the `Value` of `a` and `b` cookie