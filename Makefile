genmocks:
	mockgen -source=./db/interface.go -destination=./mocks/db/db_mock.go