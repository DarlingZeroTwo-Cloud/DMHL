import 'package:flutter/material.dart';
import '../api.dart';

class SettingsPage extends StatefulWidget {
  final DMHLApi api;
  final ValueChanged<Color> onSeed;
  const SettingsPage({super.key, required this.api, required this.onSeed});

  @override
  State<SettingsPage> createState() => _SettingsPageState();
}

class _SettingsPageState extends State<SettingsPage> {
  Map cfg = {};
  bool loading = true;

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    try {
      cfg = await widget.api.getConfig();
    } catch (_) {}
    loading = false;
    if (mounted) setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    if (loading) return const Center(child: CircularProgressIndicator());
    return Scaffold(
      backgroundColor: Colors.transparent,
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          _sectionTitle('下载源'),
          Card(
            elevation: 0,
            color: Colors.white.withValues(alpha: 0.6),
            child: SwitchListTile(
              title: const Text('使用国内镜像'),
              subtitle: const Text('BMCLAPI + 华为云，国内下载更快'),
              value: cfg['useMirror'] ?? true,
              onChanged: (v) => setState(() => cfg['useMirror'] = v),
            ),
          ),
          const SizedBox(height: 24),
          _sectionTitle('JVM 内存'),
          Card(
            elevation: 0,
            color: Colors.white.withValues(alpha: 0.6),
            child: Column(
              children: [
                ListTile(
                  title: const Text('最小堆内存'),
                  trailing: SizedBox(
                    width: 100,
                    child: TextField(
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(suffixText: 'MB'),
                      controller: TextEditingController(text: '${cfg['minMemoryMB'] ?? 1024}'),
                      onChanged: (s) => cfg['minMemoryMB'] = int.tryParse(s),
                    ),
                  ),
                ),
                ListTile(
                  title: const Text('最大堆内存'),
                  trailing: SizedBox(
                    width: 100,
                    child: TextField(
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(suffixText: 'MB'),
                      controller: TextEditingController(text: '${cfg['maxMemoryMB'] ?? 4096}'),
                      onChanged: (s) => cfg['maxMemoryMB'] = int.tryParse(s),
                    ),
                  ),
                ),
                ListTile(
                  title: const Text('垃圾回收器'),
                  trailing: DropdownButton<String>(
                    value: cfg['gcType'] ?? 'G1',
                    underline: const SizedBox(),
                    items: const [
                      DropdownMenuItem(value: 'G1', child: Text('G1（推荐）')),
                      DropdownMenuItem(value: 'ZGC', child: Text('ZGC（低延迟）')),
                    ],
                    onChanged: (v) => setState(() => cfg['gcType'] = v),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),
          _sectionTitle('外观主题'),
          Card(
            elevation: 0,
            color: Colors.white.withValues(alpha: 0.6),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('主题色'),
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 12,
                    children: [
                      for (final c in [
                        const Color(0xFF6C8CFF),
                        const Color(0xFFFF6C8C),
                        const Color(0xFF6CFFB8),
                        const Color(0xFFFFC86C),
                        const Color(0xFFB86CFF),
                        const Color(0xFF00C9B7),
                      ])
                        GestureDetector(
                          onTap: () => widget.onSeed(c),
                          child: CircleAvatar(backgroundColor: c, radius: 20),
                        ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 32),
          Center(
            child: OutlinedButton.icon(
              onPressed: () async {
                final r = await widget.api.checkUpdate();
                if (!mounted) return;
                final gh = r['github'];
                final latest = gh != null ? gh['tag_name'] : '未知';
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text('当前 v${r['current']}，最新 $latest')),
                );
              },
              icon: const Icon(Icons.system_update),
              label: const Text('检查 DMHL 更新'),
            ),
          ),
          const SizedBox(height: 12),
          Center(
            child: FilledButton.icon(
              onPressed: () async {
                await widget.api.saveConfig(Map<String, dynamic>.from(cfg));
                if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('已保存')));
                }
              },
              icon: const Icon(Icons.save),
              label: const Text('保存设置'),
            ),
          ),
        ],
      ),
    );
  }

  Widget _sectionTitle(String t) => Padding(
        padding: const EdgeInsets.only(bottom: 8, left: 4),
        child: Text(t, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
      );
}
