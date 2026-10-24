import 'package:flutter/material.dart';
import '../api.dart';

class LogsPage extends StatefulWidget {
  final DMHLApi api;
  final String instanceId;
  const LogsPage({super.key, required this.api, required this.instanceId});

  @override
  State<LogsPage> createState() => _LogsPageState();
}

class _LogsPageState extends State<LogsPage> {
  String log = '';

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    log = await widget.api.getLogs(widget.instanceId);
    if (mounted) setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('启动日志'), actions: [
        IconButton(onPressed: load, icon: const Icon(Icons.refresh)),
      ]),
      body: Container(
        color: const Color(0xFF1e1e1e),
        padding: const EdgeInsets.all(12),
        child: SingleChildScrollView(
          child: SelectableText(
            log.isEmpty ? '暂无日志' : log,
            style: const TextStyle(fontFamily: 'monospace', fontSize: 12, color: Color(0xFFd4d4d4)),
          ),
        ),
      ),
    );
  }
}
