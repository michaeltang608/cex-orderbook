dex_match_orderbook:
	go run cmd/match/main.go cmd/match/wire_gen.go -f conf/config.yaml
debug:
	go run cmd/match/main.go cmd/match/wire_gen.go -f conf/local.yaml