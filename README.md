# JSON server
This is a ready-to-use back-end server. If you want to write any client-side program but you don't want to write a back-end yet, you can use this one to try quickly.
You can change json filename (`-db`) and listening port (`-port`) with command line options.

It's really easy to use: 

`GET /table-name` returns all rows in table-name

`GET /table-name/1` returns second row in table-name

`POST /table-name` adds json body as a new row in table-name

`PUT /table-name/1` modifies second row with json body in table-name

`DELETE /table-name` deletes table-name from database

`DELETE /table-name/1` deletes second row in table-name

Moreover, some verbose errors are shown if you do something wrong.
