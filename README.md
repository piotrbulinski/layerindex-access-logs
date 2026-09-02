# Layer Index Web Log Analyzer

This project will import public web server logs from the official instance of the layer index web,
parse them and put them into a postgresql database for exploration with Metabase.

## Importing data

Run:

```shell
# start the docker compose services
docker compose up -d

# import the data
go run main.go
```

and wait.


## Accessing the imported data

Visit http://localhost:3000/, sign-up and enjoy access to the data!