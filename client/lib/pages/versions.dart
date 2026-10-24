import 'package:flutter/material.dart';
import '../api.dart';

class VersionsPage extends StatefulWidget {
  final DMHLApi api;
  const VersionsPage({super.key, required this.api});

  @override
  State<VersionsPage> createState() => _VersionsPageState();
}

class _VersionsPageState extends State<VersionsPage> {
  List versions = [];
  List filtered = [];
  String filter = 'release';
  bool loading = true;
  bool installing = false;
  String keyword = '';
  String? installingVer;

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    setState(() => loading = true);
    try {
      versions = await widget.api.versions();
    } catch (e) {}
    applyFilter();
    setState(() => loading = false);
  }

  void applyFilter() {
    filtered = versions.where((v) {
      if (filter != 'all' && v['type'] != filter) return false;
      if (keyword.isNotEmpty && !(v['id'] as String).contains(keyword)) return false;
      return true;
    }).toList();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Column(
        children: [
          // 顶部搜索栏
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    decoration: InputDecoration(
                      prefixIcon: const Icon(Icons.search),
                      hintText: '搜索版本号...',
                      filled: true,
                      fillColor: Colors.white.withValues(alpha: 0.5),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
                    ),
                    onChanged: (s) {
                      keyword = s;
                      applyFilter();
                    },
                  ),
                ),
                const SizedBox(width: 12),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.5),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: DropdownButton<String>(
                    value: filter,
                    underline: const SizedBox(),
                    items: const [
                      DropdownMenuItem(value: 'all', child: Text('全部')),
                      DropdownMenuItem(value: 'release', child: Text('正式版')),
                      DropdownMenuItem(value: 'snapshot', child: Text('快照')),
                      DropdownMenuItem(value: 'old_beta', child: Text('远古 Beta')),
                      DropdownMenuItem(value: 'old_alpha', child: Text('远古 Alpha')),
                    ],
                    onChanged: (v) => setState(() {
                      filter = v!;
                      applyFilter();
                    }),
                  ),
                ),
              ],
            ),
          ),
          // 安装进度条
          if (installing)
            Container(
              margin: const EdgeInsets.symmetric(horizontal: 16),
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.primaryContainer.withValues(alpha: 0.5),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Row(
                children: [
                  const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)),
                  const SizedBox(width: 12),
                  Expanded(child: Text('正在安装 $installingVer ...')),
                ],
              ),
            ),
          Expanded(
            child: loading
                ? const Center(child: CircularProgressIndicator())
                : ListView.separated(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    itemCount: filtered.length,
                    separatorBuilder: (_, __) => const Divider(height: 1),
                    itemBuilder: (_, i) {
                      final v = filtered[i];
                      final isInstalling = installingVer == v['id'];
                      return ListTile(
                        leading: Icon(_typeIcon(v['type']), color: _typeColor(v['type'])),
                        title: Text(v['id'], style: const TextStyle(fontWeight: FontWeight.w600)),
                        subtitle: Text('${v['type']} · ${(v['releaseTime'] ?? '').toString().substring(0, 10)}'),
                        trailing: FilledButton(
                          onPressed: installing ? null : () => _install(v['id']),
                          child: isInstalling
                              ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                              : const Text('下载'),
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  IconData _typeIcon(String? t) {
    switch (t) {
      case 'release': return Icons.verified;
      case 'snapshot': return Icons.bug_report;
      case 'old_beta': return Icons.history;
      case 'old_alpha': return Icons.history_toggle_off;
      default: return Icons.help;
    }
  }

  Color _typeColor(String? t) {
    switch (t) {
      case 'release': return Colors.green;
      case 'snapshot': return Colors.orange;
      default: return Colors.grey;
    }
  }

  Future<void> _install(String id) async {
    // 先创建实例再安装
    setState(() {
      installing = true;
      installingVer = id;
    });
    try {
      final ins = await widget.api.createInstance({
        'name': '$id 实例',
        'versionId': id,
        'loader': 'vanilla',
      });
      final res = await widget.api.installInstance(ins['id']);
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(res['message'] ?? '安装完成')),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('安装失败: $e'), backgroundColor: Colors.red),
      );
    } finally {
      setState(() {
        installing = false;
        installingVer = null;
      });
    }
  }
}
