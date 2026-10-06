
using System;
using System.Windows;
using System.Windows.Controls;

namespace CyberSaaS.Terminal.UI.Views;

public partial class TerminalHomeView : UserControl
{
    private readonly Action _startSessionCallback;

    public TerminalHomeView(
        Action startSessionCallback)
    {
        InitializeComponent();

        _startSessionCallback =
            startSessionCallback;
    }

    private void StartSessionButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        _startSessionCallback();
    }
}
