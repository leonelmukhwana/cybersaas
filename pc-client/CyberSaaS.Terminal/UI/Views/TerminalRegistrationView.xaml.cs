using System;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;

namespace CyberSaaS.Terminal.UI.Views;

public partial class TerminalRegistrationView : UserControl
{
    private readonly Func<string, Task> _registerCallback;

    public TerminalRegistrationView(
        Func<string, Task> registerCallback)
    {
        InitializeComponent();

        _registerCallback = registerCallback;
    }

    private async void RegisterButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        string licenceKey =
            LicenceKeyTextBox.Text.Trim();

        if (string.IsNullOrWhiteSpace(licenceKey))
        {
            StatusTextBlock.Text =
                "Please enter the enrollment key.";

            return;
        }

        RegisterButton.IsEnabled = false;

        StatusTextBlock.Text =
            "🟡 Registering computer...";

        try
        {
            await _registerCallback(licenceKey);
        }
        catch (Exception ex)
        {
            StatusTextBlock.Text =
                $"🔴 Registration failed.\n\n{ex.Message}";
        }
        finally
        {
            RegisterButton.IsEnabled = true;
        }
    }
}