import 'package:flutter/material.dart';
import '../api.dart';

class AccountPage extends StatefulWidget {
  final DMHLApi api;
  const AccountPage({super.key, required this.api});

  @override
  State<AccountPage> createState() => _AccountPageState();
}

class _AccountPageState extends State<AccountPage> {
  List accounts = [];
  String nameCtl = '';

  @override
  void initState() {
    super.initState();
    reload();
  }

  Future<void> reload() async {
    accounts = await widget.api.accounts();
    if (mounted) setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          const Text('账号登录', style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold)),
          const SizedBox(height: 16),
          _card(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('离线登录', style: TextStyle(fontWeight: FontWeight.bold)),
                const SizedBox(height: 12),
                TextField(
                  decoration: const InputDecoration(hintText: '输入玩家名', border: OutlineInputBorder()),
                  onChanged: (s) => nameCtl = s,
                ),
                const SizedBox(height: 12),
                FilledButton.icon(
                  onPressed: () async {
                    if (nameCtl.isEmpty) return;
                    await widget.api.offlineLogin(nameCtl);
                    reload();
                  },
                  icon: const Icon(Icons.login),
                  label: const Text('添加离线账号'),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),
          _card(
            child: ListTile(
              leading: const Icon(Icons.account_circle, size: 32),
              title: const Text('微软登录'),
              subtitle: const Text('Microsoft 正版账号（将打开浏览器授权）'),
              trailing: FilledButton(onPressed: _microsoftLogin, child: const Text('登录')),
            ),
          ),
          const SizedBox(height: 12),
          _card(
            child: ListTile(
              leading: const Icon(Icons.public, size: 32),
              title: const Text('LittleSkin 外置登录'),
              subtitle: const Text('国内第三方皮肤站'),
              trailing: FilledButton(onPressed: _littleSkinDialog, child: const Text('登录')),
            ),
          ),
          const SizedBox(height: 24),
          const Text('已添加账号', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
          const SizedBox(height: 8),
          ...accounts.map((a) => _card(
                child: ListTile(
                  leading: const CircleAvatar(child: Icon(Icons.person)),
                  title: Text(a['username']),
                  subtitle: Text('${a['type']} · ${a['uuid']}'),
                  trailing: IconButton(
                    icon: const Icon(Icons.image_outlined),
                    tooltip: '导入皮肤 PNG',
                    onPressed: () => _importSkin(a['uuid']),
                  ),
                ),
              )),
        ],
      ),
    );
  }

  Widget _card({required Widget child}) => Card(
        elevation: 0,
        color: Colors.white.withValues(alpha: 0.6),
        child: Padding(padding: const EdgeInsets.all(12), child: child),
      );

  void _todo(String name) {
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('$name 将在后续版本实现')));
  }

  Future<void> _microsoftLogin() async {
    try {
      final r = await widget.api.get('/api/auth/microsoft/url');
      final url = r['url'] as String;
      if (!mounted) return;
      showDialog(
        context: context,
        builder: (_) => AlertDialog(
          title: const Text('微软登录'),
          content: SelectableText('请在浏览器中打开：\n\n$url\n\n授权后此窗口会自动收到回调。'),
          actions: [TextButton(onPressed: () => Navigator.pop(context), child: const Text('好'))],
        ),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('错误: $e')));
    }
  }

  void _littleSkinDialog() {
    final baseCtl = TextEditingController(text: 'https://littleskin.cn/api/yggdrasil');
    final userCtl = TextEditingController();
    final passCtl = TextEditingController();
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('LittleSkin 登录'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: baseCtl, decoration: const InputDecoration(labelText: 'Base URL')),
            TextField(controller: userCtl, decoration: const InputDecoration(labelText: '用户名/邮箱')),
            TextField(controller: passCtl, decoration: const InputDecoration(labelText: '密码'), obscureText: true),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('取消')),
          FilledButton(
            onPressed: () async {
              try {
                await widget.api.post('/api/auth/littleskin', {
                  'baseUrl': baseCtl.text,
                  'username': userCtl.text,
                  'password': passCtl.text,
                });
                if (!mounted) return;
                Navigator.pop(context);
                reload();
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('LittleSkin 登录成功')));
              } catch (e) {
                ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('登录失败: $e')));
              }
            },
            child: const Text('登录'),
          ),
        ],
      ),
    );
  }

  void _importSkin(String uuid) {
    final pathCtl = TextEditingController();
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('导入皮肤 PNG'),
        content: TextField(
          controller: pathCtl,
          decoration: const InputDecoration(hintText: '/path/to/skin.png'),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('取消')),
          FilledButton(
            onPressed: () async {
              if (pathCtl.text.isEmpty) return;
              final r = await widget.api.importSkin(pathCtl.text, uuid);
              if (mounted) {
                Navigator.pop(context);
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text(r['ok'] == true ? '皮肤导入成功' : '导入失败')),
                );
              }
            },
            child: const Text('导入'),
          ),
        ],
      ),
    );
  }
}
