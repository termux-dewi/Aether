.PHONY: build build-server build-client test clean proto vet \
	gomobile-bind app-install app-typecheck app-prebuild app-build-apk app-build-apk-cloud

GOBIN := bin

build: build-server build-client

build-server:
	go build -o $(GOBIN)/aether-server ./cmd/aether-server

build-client:
	go build -o $(GOBIN)/aether-client ./cmd/aether-client

test:
	go test ./pkg/... ./internal/... -count=1

test-verbose:
	go test ./pkg/... ./internal/... -v -count=1

test-race:
	go test ./pkg/... ./internal/... -race -count=1

vet:
	go vet ./...

proto:
	bash scripts/generate_proto.sh

clean:
	rm -rf $(GOBIN)

# Generate gomobile bindings for Android (ARM only — no x86/x86_64).
# In CI this is produced by .github/workflows/build-aar.yml and downloaded
# as the `aether-aar` artifact; locally run this target with gomobile + NDK.
gomobile-bind:
	gomobile bind -target=android/arm,android/arm64 \
		-androidapi 26 \
		-o android/AetherBox/app/libs/aether.aar \
		./mobile

# --- React Native Android client (app/) ---

app-install:
	cd app && npm install

app-typecheck:
	cd app && npx tsc --noEmit

# Generate the native android/ folder from app.json (expo prebuild)
app-prebuild:
	cd app && npx expo prebuild --platform android --clean

# Build APK locally via EAS CLI (requires Android SDK + JDK 17)
app-build-apk:
	cd app && npx eas build --platform android --profile preview --local

# Build APK on EAS cloud (requires Expo account login)
app-build-apk-cloud:
	cd app && npx eas build --platform android --profile preview