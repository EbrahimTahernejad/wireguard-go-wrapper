apple:
	@gomobile bind -o ./build/WG.xcframework -target=ios,macos,iossimulator -ldflags="-s -w" -v ./
ios:
	@gomobile bind -o ./build/ios/WG.xcframework -target=ios,iossimulator -ldflags="-s -w" -v ./
macos:
	@gomobile bind -o ./build/macos/WG.xcframework -target=macos -ldflags="-s -w" -v ./
