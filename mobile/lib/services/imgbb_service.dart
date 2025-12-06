import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;

class ImgBBService {
  // Free ImgBB API key (you can get your own at imgbb.com)
  static const String _apiKey = '6d207e02198a847aa98d0a2a901485a5';
  static const String _uploadUrl = 'https://api.imgbb.com/1/upload';

  /// Upload image to ImgBB and return the URL
  static Future<String?> uploadImage(File imageFile) async {
    try {
      final bytes = await imageFile.readAsBytes();
      final base64Image = base64Encode(bytes);

      final response = await http.post(
        Uri.parse(_uploadUrl),
        body: {
          'key': _apiKey,
          'image': base64Image,
        },
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        if (data['success'] == true) {
          return data['data']['url'] as String;
        }
      }
      
      print('[ImgBBService] Upload failed: ${response.body}');
      return null;
    } catch (e) {
      print('[ImgBBService] Error: $e');
      return null;
    }
  }

  /// Upload image from path
  static Future<String?> uploadImageFromPath(String path) async {
    final file = File(path);
    if (!await file.exists()) {
      print('[ImgBBService] File not found: $path');
      return null;
    }
    return uploadImage(file);
  }
}

