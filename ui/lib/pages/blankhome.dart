import "package:flutter/material.dart";
import "package:url_launcher/url_launcher.dart";

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
              const ElevatedButton(
                onPressed: googleLogin, child: Text("Add Google Account"),
              )
            ],
          ),
        ),
      ),
    );
  }
}

final Uri _url = Uri.parse("https://flutter.dev");

Future<void> googleLogin() async {
  if (!await launchUrl(_url)) {
    throw Exception("Could not launch $_url");
  }
}