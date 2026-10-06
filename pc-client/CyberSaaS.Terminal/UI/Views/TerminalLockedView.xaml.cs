using System;
using System.Windows;
using System.Windows.Controls;

namespace CyberSaaS.Terminal.UI.Views;

public partial class TerminalLockedView : UserControl
{
    private readonly Action _startVerificationCallback;

    public TerminalLockedView(
        Action startVerificationCallback)
    {
        InitializeComponent();

        _startVerificationCallback =
            startVerificationCallback
            ?? throw new ArgumentNullException(
                nameof(startVerificationCallback));
    }

    private void StartVerificationButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        _startVerificationCallback();
    }
}