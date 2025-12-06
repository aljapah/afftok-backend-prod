# Flutter specific rules
-keep class io.flutter.app.** { *; }
-keep class io.flutter.plugin.** { *; }
-keep class io.flutter.util.** { *; }
-keep class io.flutter.view.** { *; }
-keep class io.flutter.** { *; }
-keep class io.flutter.plugins.** { *; }

# Keep Google Sign-In
-keep class com.google.android.gms.** { *; }

# Keep http package
-keep class okhttp3.** { *; }
-keep class okio.** { *; }

# General rules
-keepattributes *Annotation*
-keepattributes Signature
-keepattributes Exceptions

