import 'package:flutter/material.dart';
import '../api.dart';

class ModsPage extends StatefulWidget {
  final DMHLApi api;
  const ModsPage({super.key, required this.api});

  @override
  State<ModsPage> createState() => _ModsPageState();
}

class _ModsPageState extends State<ModsPage> {
  List hits = [];
  bool loading = false;
  String keyword = 'sodium';

  @override
  void initState() {
    super.initState();
    search();
  }

  Future<void> search() async {
    setState(() => loading = true);
    try {
      hits = await widget.api.modrinthSearch(keyword);
    } catch (e) {
      hits = [];
    }
    setState(() => loading = false);
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
                    decoration: InputDecoration(
                      prefixIcon: const Icon(Icons.search),
                      hintText: '搜索 Modrinth mod...',
                      filled: true,
                      fillColor: Colors.white.withValues(alpha: 0.5),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
                    ),
                    onSubmitted: (s) {
                      keyword = s;
                      search();
                    },
                  ),
                ),
                const SizedBox(width: 12),
                FilledButton(onPressed: search, child: const Text('搜索')),
              ],
            ),
          ),
          Expanded(
            child: loading
                ? const Center(child: CircularProgressIndicator())
                : ListView.separated(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    itemCount: hits.length,
                    separatorBuilder: (_, __) => const SizedBox(height: 8),
                    itemBuilder: (_, i) {
                      final m = hits[i];
                      return Card(
                        elevation: 0,
                        color: Colors.white.withValues(alpha: 0.6),
                        child: ListTile(
                          leading: ClipRRect(
                            borderRadius: BorderRadius.circular(8),
                            child: m['icon_url'] != null && m['icon_url'].toString().isNotEmpty
                                ? Image.network(m['icon_url'], width: 48, height: 48, fit: BoxFit.cover,
                                    errorBuilder: (_, __, ___) => const Icon(Icons.extension, size: 48))
                                : const Icon(Icons.extension, size: 48),
                          ),
                          title: Text(m['title'], style: const TextStyle(fontWeight: FontWeight.bold)),
                          subtitle: Text(m['description'] ?? '', maxLines: 2, overflow: TextOverflow.ellipsis),
                          trailing: FilledButton.icon(
                            icon: const Icon(Icons.download),
                            label: Text('${_fmt(m['downloads'])}'),
                            onPressed: () => _downloadMod(m),
                          ),
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  String _fmt(n) {
    if (n == null) return '0';
    if (n >= 1000000) return '${(n / 1000000).toStringAsFixed(1)}M';
    if (n >= 1000) return '${(n / 1000).toStringAsFixed(1)}K';
    return '$n';
  }

  Future<void> _downloadMod(Map m) async {
    // 先列实例让用户选
    final insts = await widget.api.instances();
    if (!mounted) return;
    if (insts.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('请先创建一个实例')));
      return;
    }
    String? pickedId;
    await showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: Text('下载 ${m['title']} 到哪个实例？'),
        content: SizedBox(
          width: 300,
          child: ListView(
            shrinkWrap: true,
            children: insts.map<Widget>((i) => ListTile(
                  title: Text(i['name'] ?? i['id']),
                  subtitle: Text(i['versionId'] ?? ''),
                  onTap: () {
                    pickedId = i['id'];
                    Navigator.pop(context);
                  },
                )).toList(),
          ),
        ),
      ),
    );
    if (pickedId == null || !mounted) return;
    final r = await widget.api.downloadMod(m['slug'], '', 'fabric', pickedId!);
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(r['message'] ?? '下载完成')),
    );
  }
}
