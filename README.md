##KRanker (2014)...


This repo contains code to extract Kenya Certificate of Primary Examination results from the Kenya National Examination Council website in 2014.

The data was meant to be used to perform ad hoc student and school analysis and keep people honest since the government wasn't willing to provide analysis on the data.

The website is down now, but it was fun to use this to get a dump of the results. There was no time do analysis and the data was deleted.

The code contains great golang related nuggets: go routines, empty structs as boolean checks, timeouts for requests and back off techniques. 


The output is a CSV of candidate results:

```
$ go get github.com/sgichohi/kranker

```

```

$ kranker

total,eng,kis,mat,sci,ssr,schoolName,gender,engGrade,kisGrade,matGrade,sciGrade,ssrGrade
291,61,53,66,69,42,YYYYY,F,B-,C,B,B,C-
...
...
...

```
