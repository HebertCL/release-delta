# Release reporter

Query releases using semantic versioning from open source repos

## Usage

- `make tidy` - Install dependencies
- `make build` - Build Docker image
- `make start-local` - Run program locally using Go

## References and downloads

```bash
docker pull hebertcuellar/release-reporter:latest
```

## TODOs and future improvements

- Handle GitHub authentication
- Perform tag validation to avoid unexpected results which do not fail by default
- Ability to parametrize values such as execution port