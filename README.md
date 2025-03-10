# mypod

Start your own podcast with just a folder of files.

[![Go Reference](https://pkg.go.dev/badge/github.com/parkr/mypod.svg)](https://pkg.go.dev/github.com/parkr/mypod)

### Example configuration

The podcast-wide configuration can be placed in `podcast.json` in your storage directory:

```json
{
  "BaseURL": "https://mypod.yourdomain.com",
  "Title": "My Personal Podcast",
  "Link": "https://mypod.yourdomain.com/",
  "Description": "A podcast of my favorite things.",
  "Language": "EN",
  "Copyright": "N/A",
  "Author": "Various",
  "Subtitle": "A subtitle",
  "Summary": "A summary",
  "Owner": {
    "Name": "Example Person",
    "Email": "you@example.com"
  },
  "Image": "/podcast.jpg",
  "Categories": ["Business", "News", "Technology"],
  "Explicit": false
}
```

- `Link` – The HTML URL you want a visitor to see when they tap on the podcast URL
- `Image` – The URL path to your podcast's image file (without the BaseURL)

### MIME Types

Go builds in only a small number of MIME types, preferring instead to read the
`mime.types` files commonly found on disk. On Unix, this includes `/etc/mime.types`,
`/etc/apache2/mime.types`, `/etc/apache/mime.types`, and `/etc/httpd/conf/mime.types`.
When running mypod in the provided Docker image, MIME types are downloaded from the Apache2
project. If you have custom MIME types on your server, mount them as a read-only volume to one of these paths.
**Note:** MIME types must be in the format `<type> <ext>`, with no semi-colons or any other text.
For example, the nginx `mime.types` file is an nginx directive and not a strict MIME types file
so the Go MIME parsing will fail to read this properly.

## Download Customization

### Cookies

Provide cookies info in `yt-dl-cookies.txt` in your storage dir.

### Additional yt-dlp flags

If you'd like to customize further the `yt-dlp` command that's run to download
the podcast episode, place them into an array in `yt-dl-args.json` in your storage dir.
The array must contain only strings -- any numerics or anything will fail.
