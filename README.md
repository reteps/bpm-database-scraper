# bpm-database-scraper

### No longer in use, if you find this also take a look at `tempo` search of echo nest api.
- https://echonest.github.io/pyechonest/
Language
- Golang for concurrent requests and high speed
Database
- PostgreSQL
  - Offers Golang support
  - I have mostly structured data
  - https://www.alooma.com/blog/types-of-modern-databases
  - Totaled `79409` songs with this program
Thread Limit
- `ulimit -n 10000`
Environment
- Have PostgreSQL db running in background
- .env file containing `USER` and `PASSWORD`
