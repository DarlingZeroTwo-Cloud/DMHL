import 'package:flutter/material.dart';
import '../api.dart';

class HomePage extends StatefulWidget {
  final DMHLApi api;
  const HomePage({super.key, required this.api});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  List instances = [];
  bool loading = true;
  String status = '';

  @override
  void initState() {
    super.initState();
    reload();
  }

  Future<void> reload() async {
    setState(() => loading = true);
    try {
      instances = await widget.api.instances();
    } catch (e) {
      status = '后端未启动: $e';
    }
    setState(() => loading = false);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showCreateDialog(),
        icon: const Icon(Icons.add),
        label: const Text('新建实例'),
      ),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : instances.isEmpty
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.inbox, size: 64, color: Colors.grey),
                      const SizedBox(height: 12),
                      const Text('还没有实例，点右下角新建一个'),
                      if (status.isNotEmpty) Text(status, style: const TextStyle(color: Colors.red)),
                    ],
                  ),
                )
              : ListView.separated(
                  padding: const EdgeInsets.all(24),
                  itemCount: instances.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 12),
                  itemBuilder: (_, i) {
                    final ins = instances[i];
                    return Card(
                      child: ListTile(
                        leading: const Icon(Icons.sports_esports, size: 36),
                        title: Text(ins['name'] ?? ins['id']),
                        subtitle: Text('版本 ${ins['versionId']} · ${ins['loader'] ?? 'vanilla'}'),
                        trailing: FilledButton(
                          onPressed: () => _launch(ins['id']),
                          child: const Text('启动'),
                        ),
                      ),
                    );
                  },
                ),
    );
  }

  Future<void> _launch(String id) async {
    final r = await widget.api.launch(id);
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(r['message'] ?? '已发送启动请求')),
    );
  }

  void _showCreateDialog() {
    final nameCtl = TextEditingController();
    final verCtl = TextEditingController(text: '1.21.4');
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('新建实例'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: nameCtl, decoration: const InputDecoration(labelText: '实例名')),
            const SizedBox(height: 12),
            TextField(controller: verCtl, decoration: const InputDecoration(labelText: 'MC 版本号')),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('取消')),
          FilledButton(
            onPressed: () async {
              await widget.api.createInstance({
                'name': nameCtl.text,
                'versionId': verCtl.text,
                'loader': 'vanilla',
              });
              if (mounted) Navigator.pop(context);
              reload();
            },
            child: const Text('创建'),
          ),
        ],
      ),
    );
  }
}
