import "package:flutter/material.dart";
import "package:url_launcher/url_launcher.dart";
import 'package:webview_flutter/webview_flutter.dart';
import 'package:logging/logging.dart';

final log = Logger("BlankHome");

class BlankHome extends StatefulWidget {
  const BlankHome({super.key});

  @override
  State<BlankHome> createState() => _BlankHomeState();
}

class _BlankHomeState extends State<BlankHome> {
  // final _formKey = GlobalKey<FormState>();
  // final _usernameController = TextEditingController();
  // final _passwordController = TextEditingController();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: SingleChildScrollView(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Image.asset(
                "assets/images/gmail.png",
                width: 100,
                height: 100,
              ), // Optional logo
              const SizedBox(height: 20),
              ElevatedButton(
                onPressed: googleLogin,
                child: Text("Add Google Account"),
              )
            ],
          ),
        ),
      ),
    );
  }
}

final Uri googleLoginApi = Uri.parse("http://localhost:59775/google/login");

Future<void> googleLogin() async {
  if (!await launchUrl(googleLoginApi, mode: LaunchMode.externalApplication)) {
    throw Exception("Could not launch $googleLoginApi");
  }
  log.info("hello");
}