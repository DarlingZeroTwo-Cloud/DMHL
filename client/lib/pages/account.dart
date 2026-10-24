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
              subtitle: const Text('完整 OAuth 流程（后续版本）'),
              trailing: FilledButton(onPressed: () => _todo('微软登录'), child: const Text('登录')),
            ),
          ),
          const SizedBox(height: 12),
          _card(
            child: ListTile(
              leading: const Icon(Icons.public, size: 32),
              title: const Text('LittleSkin 外置登录'),
              subtitle: const Text('国内第三方皮肤站（后续版本）'),
              trailing: FilledButton(onPressed: () => _todo('LittleSkin'), child: const Text('配置')),
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
