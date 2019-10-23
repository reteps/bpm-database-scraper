# bpm-database-scraper

Scapes the website [bpmdatabase.com](http://bpmdatabase.com) For every songs BPM
## Language
- Golang for concurrent requests and high speed
## Database
- PostgreSQL
  - Offers Golang support
  - I have mostly structured data
  
## Notes
#### Change the thread limit
- `ulimit -n 10000`
#### Environment
- Have PostgreSQL db running in background
  - See main.go for commands I ran to initialize table
- .env file containing `USER` and `PASSWORD`
## Success

- Totaled `79409` songs with this program
