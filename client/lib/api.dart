import 'dart:convert';
import 'package:http/http.dart' as http;

/// DMHL Go 后端 API 客户端
class DMHLApi {
  final String base;
  DMHLApi({this.base = 'http://127.0.0.1:18923'});

  Future<Map<String, dynamic>> health() async {
    final r = await http.get(Uri.parse('$base/api/health'));
    return jsonDecode(r.body);
  }

  Future<List<dynamic>> versions() async {
    final r = await http.get(Uri.parse('$base/api/versions'));
    return jsonDecode(r.body)['versions'] ?? [];
  }

  Future<List<dynamic>> instances() async {
    final r = await http.get(Uri.parse('$base/api/instances'));
    return jsonDecode(r.body);
  }

  Future<Map<String, dynamic>> createInstance(Map<String, dynamic> body) async {
    final r = await http.post(
      Uri.parse('$base/api/instances/create'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode(body),
    );
    return jsonDecode(r.body);
  }

  Future<Map<String, dynamic>> launch(String id) async {
    final r = await http.post(
      Uri.parse('$base/api/instances/launch'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'id': id}),
    );
    return jsonDecode(r.body);
  }

  Future<Map<String, dynamic>> offlineLogin(String username) async {
    final r = await http.post(
      Uri.parse('$base/api/auth/offline'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'username': username}),
    );
    return jsonDecode(r.body);
  }

  Future<List<dynamic>> accounts() async {
    final r = await http.get(Uri.parse('$base/api/accounts'));
    return jsonDecode(r.body);
  }

  Future<Map<String, dynamic>> getConfig() async {
    final r = await http.get(Uri.parse('$base/api/config'));
    return jsonDecode(r.body);
  }

  Future<Map<String, dynamic>> saveConfig(Map<String, dynamic> cfg) async {
    final r = await http.put(
      Uri.parse('$base/api/config'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode(cfg),
    );
    return jsonDecode(r.body);
  }
}
