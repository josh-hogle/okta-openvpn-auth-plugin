BUILD_DIR=build
VERSION=1.0.0

all: $(BUILD_DIR)/okta-openvpn $(BUILD_DIR)/auth_script.so

${BUILD_DIR}:
	mkdir -p ${BUILD_DIR}

$(BUILD_DIR)/auth_script.so: ${BUILD_DIR}
	make -C third_party/auth-script-openvpn
	mv third_party/auth-script-openvpn/auth_script.so $(BUILD_DIR)

$(BUILD_DIR)/okta-openvpn: cmd/okta-openvpn/*.go
	GO111MODULE=on go build -ldflags="-X main.version=${VERSION} -s -w" -o $(BUILD_DIR)/okta-openvpn ./cmd/okta-openvpn

clean:
	rm -rf $(BUILD_DIR)
