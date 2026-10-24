import 'package:flutter/material.dart';
import 'api.dart';
import 'theme.dart';
import 'pages/home.dart';
import 'pages/versions.dart';
import 'pages/settings.dart';
import 'pages/account.dart';

void main() {
  runApp(const DMHLApp());
}

class DMHLApp extends StatefulWidget {
  const DMHLApp({super.key});

  @override
  State<DMHLApp> createState() => _DMHLAppState();
}

class _DMHLAppState extends State<DMHLApp> {
  Color seed = const Color(0xFF6C8CFF);
  int index = 0;
  final api = DMHLApi();

  @override
  Widget build(BuildContext context) {
    final isDark = MediaQuery.platformBrightnessOf(context) == Brightness.dark;
    return MaterialApp(
      title: 'DMHL',
      debugShowCheckedModeBanner: false,
      themeMode: ThemeMode.system,
      theme: DMHLTheme.light(seed),
      darkTheme: DMHLTheme.dark(seed),
      home: Scaffold(
        body: Container(
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: isDark
                  ? [const Color(0xFF1a1a2e), const Color(0xFF16213e)]
                  : [const Color(0xFFe0e7ff), const Color(0xFFf0f4ff)],
            ),
          ),
          child: Row(
            children: [
              NavigationRail(
                extended: MediaQuery.sizeOf(context).width > 900,
                backgroundColor: Colors.transparent,
                selectedIndex: index,
                onDestinationSelected: (i) => setState(() => index = i),
                destinations: const [
                  NavigationRailDestination(icon: Icon(Icons.home), label: Text('实例')),
                  NavigationRailDestination(icon: Icon(Icons.download), label: Text('版本下载')),
                  NavigationRailDestination(icon: Icon(Icons.person), label: Text('账号皮肤')),
                  NavigationRailDestination(icon: Icon(Icons.settings), label: Text('设置')),
                ],
              ),
              const VerticalDivider(width: 1),
              Expanded(
                child: ClipRRect(
                  borderRadius: const BorderRadius.only(
                    topLeft: Radius.circular(20),
                    bottomLeft: Radius.circular(20),
                  ),
                  child: switch (index) {
                    0 => HomePage(api: api),
                    1 => VersionsPage(api: api),
                    2 => AccountPage(api: api),
                    _ => SettingsPage(api: api, onSeed: (c) => setState(() => seed = c)),
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
