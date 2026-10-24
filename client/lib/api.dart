import 'dart:convert';
import 'package:http/http.dart' as http';

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

  Future<Map<String, dynamic>> versionDetail(String url) async {
    final r = await http.get(Uri.parse('$base/api/version/detail?url=${Uri.encodeComponent(url)}'));
    return jsonDecode(r.body);
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

  /// 安装版本（下载 client.jar + libraries + assets）
  Future<Map<String, dynamic>> installInstance(String id) async {
    final r = await http.post(
      Uri.parse('$base/api/instances/install'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'id': id}),
    );
    return jsonDecode(r.body);
  }

  Future<Map<String, dynamic>> launch(String id, {String username = 'Player'}) async {
    final r = await http.post(
      Uri.parse('$base/api/instances/launch'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'id': id, 'username': username}),
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

  /// Fabric loader 列表
  Future<List<dynamic>> fabricLoaders(String mcVersion) async {
    final r = await http.get(Uri.parse('$base/api/fabric/loaders?mc=$mcVersion'));
    return jsonDecode(r.body);
  }

  /// Modrinth mod 搜索
  Future<List<dynamic>> modrinthSearch(String query, {String mc = '', String loader = ''}) async {
    final u = '$base/api/modrinth/search?q=${Uri.encodeComponent(query)}&mc=$mc&loader=$loader';
    final r = await http.get(Uri.parse(u));
    return jsonDecode(r.body);
  }

  /// 皮肤导入
  Future<Map<String, dynamic>> importSkin(String path, String uuid) async {
    final r = await http.post(
      Uri.parse('$base/api/skin/import'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'path': path, 'uuid': uuid}),
    );
    return jsonDecode(r.body);
  }
}
