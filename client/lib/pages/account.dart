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
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('离线登录'),
                  const SizedBox(height: 12),
                  TextField(
                    decoration: const InputDecoration(hintText: '输入玩家名'),
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
          ),
          const SizedBox(height: 12),
          Card(
            child: ListTile(
              leading: const Icon(Icons.microsoft, size: 32),
              title: const Text('微软登录'),
              subtitle: const Text('完整 OAuth 流程，需要浏览器授权'),
              trailing: FilledButton(onPressed: () {}, child: const Text('登录')),
            ),
          ),
          const SizedBox(height: 12),
          Card(
            child: ListTile(
              leading: const Icon(Icons.public, size: 32),
              title: const Text('LittleSkin 外置登录'),
              subtitle: const Text('国内第三方皮肤站，输入 BaseURL'),
              trailing: FilledButton(onPressed: () {}, child: const Text('配置')),
            ),
          ),
          const SizedBox(height: 24),
          const Text('已添加账号', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
          const SizedBox(height: 8),
          ...accounts.map((a) => Card(
                child: ListTile(
                  leading: const Icon(Icons.person),
                  title: Text(a['username']),
                  subtitle: Text('${a['type']} · ${a['uuid']}'),
                  trailing: IconButton(
                    icon: const Icon(Icons.image_outlined),
                    onPressed: () {},
                    tooltip: '导入皮肤 PNG',
                  ),
                ),
              )),
        ],
      ),
    );
  }
}
