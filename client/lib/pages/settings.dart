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
          const Text('下载源', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
          SwitchListTile(
            title: const Text('使用国内镜像 (BMCLAPI / 华为云)'),
            value: cfg['useMirror'] ?? true,
            onChanged: (v) => setState(() => cfg['useMirror'] = v),
          ),
          const Divider(),
          const Text('JVM 内存', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
          ListTile(
            title: const Text('最小堆内存 (MB)'),
            trailing: SizedBox(
              width: 80,
              child: TextField(
                keyboardType: TextInputType.number,
                controller: TextEditingController(text: '${cfg['minMemoryMB'] ?? 1024}'),
                onChanged: (s) => cfg['minMemoryMB'] = int.tryParse(s),
              ),
            ),
          ),
          ListTile(
            title: const Text('最大堆内存 (MB)'),
            trailing: SizedBox(
              width: 80,
              child: TextField(
                keyboardType: TextInputType.number,
                controller: TextEditingController(text: '${cfg['maxMemoryMB'] ?? 4096}'),
                onChanged: (s) => cfg['maxMemoryMB'] = int.tryParse(s),
              ),
            ),
          ),
          ListTile(
            title: const Text('垃圾回收器'),
            trailing: DropdownButton<String>(
              value: cfg['gcType'] ?? 'G1',
              items: const [
                DropdownMenuItem(value: 'G1', child: Text('G1')),
                DropdownMenuItem(value: 'ZGC', child: Text('ZGC')),
              ],
              onChanged: (v) => setState(() => cfg['gcType'] = v),
            ),
          ),
          const Divider(),
          const Text('个性化', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
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
              ])
                GestureDetector(
                  onTap: () => widget.onSeed(c),
                  child: CircleAvatar(backgroundColor: c, radius: 20),
                ),
            ],
          ),
          const SizedBox(height: 24),
          FilledButton(
            onPressed: () async {
              await widget.api.saveConfig(Map<String, dynamic>.from(cfg));
              if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('已保存')));
              }
            },
            child: const Text('保存设置'),
          ),
        ],
      ),
    );
  }
}
