.PHONY: recap render all clean

recap:
	go run ./cmd/recap --user dennislee928 --year 2025 --out ./web/recap_2025.json

render:
	cd web && npm install && npx playwright install chromium && npm run render

all: recap render

clean:
	rm -rf web/out web/recap_2025.json
