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
  String launchingId = '';

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
      floatingActionButton: Column(
        mainAxisAlignment: MainAxisAlignment.end,
        children: [
          FloatingActionButton.small(
            heroTag: 'import',
            onPressed: _importModpack,
            child: const Icon(Icons.folder_open),
          ),
          const SizedBox(height: 8),
          FloatingActionButton.extended(
            onPressed: () => _showCreateDialog(),
            icon: const Icon(Icons.add),
            label: const Text('新建实例'),
          ),
        ],
      ),
      body: loading
          ? const Center(child: CircularProgressIndicator())
          : instances.isEmpty
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.inbox, size: 64, color: Colors.grey.shade400),
                      const SizedBox(height: 12),
                      const Text('还没有实例，去"版本下载"页下载一个版本'),
                      if (status.isNotEmpty) Text(status, style: const TextStyle(color: Colors.red)),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: reload,
                  child: ListView.separated(
                    padding: const EdgeInsets.all(24),
                    itemCount: instances.length,
                    separatorBuilder: (_, __) => const SizedBox(height: 12),
                    itemBuilder: (_, i) {
                      final ins = instances[i];
                      final isLaunching = launchingId == ins['id'];
                      return Card(
                        elevation: 0,
                        color: Colors.white.withValues(alpha: 0.6),
                        child: ListTile(
                          contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
                          leading: Container(
                            padding: const EdgeInsets.all(10),
                            decoration: BoxDecoration(
                              color: Theme.of(context).colorScheme.primaryContainer,
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: const Icon(Icons.sports_esports),
                          ),
                          title: Text(ins['name'] ?? ins['id'], style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                          subtitle: Text('版本 ${ins['versionId']} · ${ins['loader'] ?? 'vanilla'}'),
                          trailing: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              IconButton(
                                icon: const Icon(Icons.delete_outline, color: Colors.red),
                                onPressed: () => _delete(ins['id'], ins['name']),
                              ),
                              const SizedBox(width: 8),
                              isLaunching
                                  ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
                                  : FilledButton.icon(
                                      onPressed: () => _launch(ins['id']),
                                      icon: const Icon(Icons.play_arrow),
                                      label: const Text('启动'),
                                    ),
                            ],
                          ),
                        ),
                      );
                    },
                  ),
                ),
    );
  }

  Future<void> _launch(String id) async {
    setState(() => launchingId = id);
    try {
      final r = await widget.api.launch(id);
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(r['message'] ?? '已启动')),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('启动失败: $e'), backgroundColor: Colors.red),
      );
    } finally {
      if (mounted) setState(() => launchingId = '');
    }
  }

  Future<void> _delete(String id, String? name) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('删除实例'),
        content: Text('确定删除 "$name"？游戏存档将一并删除。'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('取消')),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('删除')),
        ],
      ),
    );
    if (ok != true) return;
    await widget.api.deleteInstance(id);
    reload();
  }

  void _importModpack() {
    final pathCtl = TextEditingController();
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('导入整合包 (.mrpack)'),
        content: TextField(
          controller: pathCtl,
          decoration: const InputDecoration(hintText: '/path/to/pack.mrpack'),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('取消')),
          FilledButton(
            onPressed: () async {
              // 先选实例
              final insts = await widget.api.instances();
              if (!mounted) return;
              if (insts.isEmpty) {
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('请先创建实例')));
                return;
              }
              String? pickedId;
              await showDialog(
                context: context,
                builder: (_) => AlertDialog(
                  title: const Text('导入到哪个实例？'),
                  content: SizedBox(
                    width: 300,
                    child: ListView(
                      shrinkWrap: true,
                      children: insts.map<Widget>((i) => ListTile(
                            title: Text(i['name'] ?? i['id']),
                            onTap: () {
                              pickedId = i['id'];
                              Navigator.pop(context);
                            },
                          )).toList(),
                    ),
                  ),
                ),
              );
              if (pickedId == null) return;
              final r = await widget.api.importModpack(pathCtl.text, pickedId!);
              if (!mounted) return;
              Navigator.pop(context);
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(content: Text('已导入 ${r['mods'] ?? 0} 个 mod')),
              );
            },
            child: const Text('导入'),
          ),
        ],
      ),
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
            TextField(controller: nameCtl, decoration: const InputDecoration(labelText: '实例名', hintText: '我的生存世界')),
            const SizedBox(height: 12),
            TextField(controller: verCtl, decoration: const InputDecoration(labelText: 'MC 版本号', hintText: '1.21.4')),
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
