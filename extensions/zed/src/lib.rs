use zed_extension_api as zed;

struct SlashExtension;

impl zed::Extension for SlashExtension {
    fn new() -> Self {
        SlashExtension
    }

    fn context_server_command(
        &mut self,
        _context_server_id: &zed::ContextServerId,
        _project: &zed::Project,
    ) -> zed::Result<zed::Command> {
        Ok(zed::Command {
            command: "slash".to_string(),
            args: vec!["mcp".to_string(), "--port".to_string(), "8765".to_string()],
            env: vec![],
        })
    }

    fn context_server_configuration(
        &mut self,
        _context_server_id: &zed::ContextServerId,
        _project: &zed::Project,
    ) -> zed::Result<Option<zed::ContextServerConfiguration>> {
        Ok(Some(zed::ContextServerConfiguration {
            installation_instructions: "Install Slash: `brew install yoosuf/tap/slash` or see https://github.com/yoosuf/Slash#quick-install".to_string(),
            settings_schema: "{}".to_string(),
            default_settings: "{}".to_string(),
        }))
    }

    fn run_slash_command(
        &self,
        command: zed::SlashCommand,
        _args: Vec<String>,
        _worktree: Option<&zed::Worktree>,
    ) -> zed::Result<zed::SlashCommandOutput, String> {
        if command.name == "slash-stats" {
            let output = zed::process::Command::new("slash")
                .arg("stats")
                .output()?;
            let text = String::from_utf8_lossy(&output.stdout).to_string();
            Ok(zed::SlashCommandOutput {
                text,
                sections: vec![],
            })
        } else {
            Err(format!("unknown slash command: {}", command.name))
        }
    }
}

zed::register_extension!(SlashExtension);
