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
                          trailing: Text('${_fmt(m['downloads'])} 下载'),
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
}
