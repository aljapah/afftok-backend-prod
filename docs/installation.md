# AffTok SDK Installation Guide

This guide covers installation instructions for all AffTok SDKs.

## Table of Contents

- [Android SDK](#android-sdk)
- [iOS SDK](#ios-sdk)
- [Flutter SDK](#flutter-sdk)
- [React Native SDK](#react-native-sdk)
- [Web SDK](#web-sdk)
- [Server-to-Server](#server-to-server)

---

## Android SDK

### Requirements

- Android API 21+ (Lollipop)
- Kotlin 1.5+

### Gradle Installation

Add to your `build.gradle` (project level):

```gradle
allprojects {
    repositories {
        google()
        mavenCentral()
        maven { url 'https://jitpack.io' }
    }
}
```

Add to your `build.gradle` (app level):

```gradle
dependencies {
    implementation 'com.afftok:afftok-sdk:1.0.0'
}
```

### Maven Installation

```xml
<dependency>
    <groupId>com.afftok</groupId>
    <artifactId>afftok-sdk</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Permissions

Add to `AndroidManifest.xml`:

```xml
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
```

### ProGuard Rules

```proguard
-keep class com.afftok.** { *; }
-keepclassmembers class com.afftok.** { *; }
```

---

## iOS SDK

### Requirements

- iOS 13.0+
- Swift 5.5+
- Xcode 13.0+

### Swift Package Manager

Add to your `Package.swift`:

```swift
dependencies: [
    .package(url: "https://github.com/afftok/ios-sdk.git", from: "1.0.0")
]
```

Or in Xcode:
1. File → Add Packages
2. Enter: `https://github.com/afftok/ios-sdk.git`
3. Select version: 1.0.0

### CocoaPods

Add to your `Podfile`:

```ruby
pod 'AfftokSDK', '~> 1.0.0'
```

Then run:

```bash
pod install
```

### Carthage

Add to your `Cartfile`:

```
github "afftok/ios-sdk" ~> 1.0.0
```

Then run:

```bash
carthage update --platform iOS
```

---

## Flutter SDK

### Requirements

- Flutter 3.0+
- Dart 3.0+

### Installation

Add to your `pubspec.yaml`:

```yaml
dependencies:
  afftok: ^1.0.0
```

Then run:

```bash
flutter pub get
```

### Platform-Specific Setup

#### Android

Add to `AndroidManifest.xml`:

```xml
<uses-permission android:name="android.permission.INTERNET" />
```

#### iOS

No additional setup required.

---

## React Native SDK

### Requirements

- React Native 0.60+
- React 16.8+

### NPM Installation

```bash
npm install @afftok/react-native-sdk
```

### Yarn Installation

```bash
yarn add @afftok/react-native-sdk
```

### Peer Dependencies

```bash
npm install @react-native-async-storage/async-storage
```

### iOS Setup

```bash
cd ios && pod install
```

### Android Setup

No additional setup required.

---

## Web SDK

### CDN Installation

```html
<!-- Production (minified) -->
<script src="https://cdn.afftok.com/sdk/afftok.min.js"></script>

<!-- Development -->
<script src="https://cdn.afftok.com/sdk/afftok.js"></script>
```

### NPM Installation

```bash
npm install @afftok/web-sdk
```

### ES Module Import

```javascript
import Afftok from '@afftok/web-sdk';
```

### CommonJS Import

```javascript
const Afftok = require('@afftok/web-sdk');
```

### Direct Download

Download from [GitHub Releases](https://github.com/afftok/web-sdk/releases) and include:

```html
<script src="/path/to/afftok.min.js"></script>
```

---

## Server-to-Server

### Node.js

```bash
npm install axios
```

### PHP

No installation required (uses built-in cURL).

### Python

```bash
pip install httpx fastapi pydantic
```

### Go

No installation required (uses standard library).

---

## Verification

After installation, verify the SDK is working:

### Mobile SDKs (Android/iOS/Flutter/React Native)

```
// Check SDK is ready
console.log(Afftok.isReady());  // Should return true after init
console.log(Afftok.getVersion());  // Should return "1.0.0"
```

### Web SDK

```javascript
console.log(Afftok.isReady());  // true
console.log(Afftok.getVersion());  // "1.0.0"
```

---

## Troubleshooting

### Common Issues

1. **Network Errors**: Check internet connection and firewall settings
2. **Invalid API Key**: Verify your API key in the AffTok dashboard
3. **Build Errors**: Ensure all dependencies are installed correctly

### Getting Help

- Documentation: https://docs.afftok.com
- Email: support@afftok.com
- GitHub Issues: https://github.com/afftok/sdk/issues

