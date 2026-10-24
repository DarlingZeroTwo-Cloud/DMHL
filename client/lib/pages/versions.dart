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
  String keyword = '';

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
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    decoration: const InputDecoration(
                      prefixIcon: Icon(Icons.search),
                      hintText: '搜索版本号...',
                    ),
                    onChanged: (s) {
                      keyword = s;
                      applyFilter();
                    },
                  ),
                ),
                const SizedBox(width: 12),
                DropdownButton<String>(
                  value: filter,
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
                      return ListTile(
                        leading: Icon(_typeIcon(v['type'])),
                        title: Text(v['id']),
                        subtitle: Text('类型 ${v['type']} · ${v['releaseTime']?.substring(0, 10) ?? ''}'),
                        trailing: FilledButton(
                          onPressed: () => _install(v['id']),
                          child: const Text('下载'),
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

  void _install(String id) {
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: Text('下载 $id'),
        content: const Text('开始下载核心文件、库与资产。后续版本：实际下载进度会在此显示。'),
        actions: [
          FilledButton(onPressed: () => Navigator.pop(context), child: const Text('确定')),
        ],
      ),
    );
  }
}
