using System;
using System.Threading.Tasks;
using System.Windows;
using CyberSaaS.Terminal.Infrastructure.Api;
using CyberSaaS.Terminal.Storage;

namespace CyberSaaS.Terminal;

public partial class MpesaPaymentWindow : Window
{
    private readonly TerminalRegistration _registration;
    private readonly TerminalPaymentApi _paymentApi;
    private readonly string _sessionId;
    private readonly string _amount;

    private bool _processing;
    private bool _paymentCompleted;

    public MpesaPaymentWindow(
        TerminalRegistration registration,
        TerminalPaymentApi paymentApi,
        string sessionId,
        string amount)
    {
        InitializeComponent();

        _registration = registration;
        _paymentApi = paymentApi;
        _sessionId = sessionId;
        _amount = amount;

        AmountTextBlock.Text = $"KES {amount}";
    }

    private async void SendButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        if (_processing)
        {
            return;
        }

        string phone =
            PhoneNumberTextBox.Text.Trim();

        if (!IsValidKenyanPhone(phone))
        {
            MessageBox.Show(
                "Enter a valid Kenyan M-Pesa phone number.",
                "Invalid Phone Number",
                MessageBoxButton.OK,
                MessageBoxImage.Warning);

            PhoneNumberTextBox.Focus();

            return;
        }

        _processing = true;

        SendButton.IsEnabled = false;
        PhoneNumberTextBox.IsEnabled = false;

        StatusTextBlock.Text =
            "Sending STK Push...";

        try
        {
            /*
             * The actual STK request will go through
             * the CyberSaaS backend.
             *
             * The terminal must never contain
             * Daraja consumer credentials or passkeys.
             */

            MpesaStkResponse response =
                await _paymentApi.SendMpesaStkAsync(
                    _sessionId,
                    _amount,
                    phone);

            StatusTextBlock.Text =
                "STK Push sent successfully.";

            WaitingTextBlock.Visibility =
                Visibility.Visible;

            /*
             * IMPORTANT:
             *
             * Do NOT mark the payment successful here.
             *
             * We wait for the backend to confirm
             * the Daraja callback.
             */

            await WaitForPaymentConfirmation(
                response.PaymentId);
        }
        catch (Exception ex)
        {
            _processing = false;

            SendButton.IsEnabled = true;
            PhoneNumberTextBox.IsEnabled = true;

            WaitingTextBlock.Visibility =
                Visibility.Collapsed;

            StatusTextBlock.Text =
                "Payment could not be started.";

            MessageBox.Show(
                ex.Message,
                "M-Pesa Payment",
                MessageBoxButton.OK,
                MessageBoxImage.Error);
        }
    }

    private async Task WaitForPaymentConfirmation(
        string paymentId)
    {
        /*
         * This will poll the CyberSaaS backend.
         *
         * The backend becomes the source of truth:
         *
         * pending
         * confirmed
         * failed
         * cancelled
         *
         * We will connect this to the terminal
         * payment-status endpoint next.
         */

        for (int attempt = 0; attempt < 60; attempt++)
        {
            await Task.Delay(
                TimeSpan.FromSeconds(2));

            MpesaPaymentStatusResponse status =
                await _paymentApi
                    .GetMpesaPaymentStatusAsync(
                        paymentId);

            if (status.Status.Equals(
                    "confirmed",
                    StringComparison.OrdinalIgnoreCase))
            {
                _paymentCompleted = true;

                StatusTextBlock.Text =
                    "Payment confirmed.";

                WaitingTextBlock.Text =
                    "Payment received successfully.";

                await Task.Delay(700);

                Close();

                return;
            }

            if (status.Status.Equals(
                    "failed",
                    StringComparison.OrdinalIgnoreCase) ||
                status.Status.Equals(
                    "cancelled",
                    StringComparison.OrdinalIgnoreCase))
            {
                throw new InvalidOperationException(
                    status.Message ??
                    "The M-Pesa payment was not completed.");
            }
        }

        throw new TimeoutException(
            "The M-Pesa payment confirmation timed out. " +
            "Check the payment status before trying again.");
    }

    private void BackButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        if (_processing)
        {
            MessageBox.Show(
                "Please wait for the current M-Pesa request to finish.",
                "Payment In Progress",
                MessageBoxButton.OK,
                MessageBoxImage.Information);

            return;
        }

        Close();
    }

    private static bool IsValidKenyanPhone(
        string phone)
    {
        phone = phone.Replace(" ", "");

        if (phone.StartsWith("+254"))
        {
            return phone.Length == 13;
        }

        if (phone.StartsWith("254"))
        {
            return phone.Length == 12;
        }

        if (phone.StartsWith("07") ||
            phone.StartsWith("01"))
        {
            return phone.Length == 10;
        }

        return false;
    }

    protected override void OnClosed(
        EventArgs e)
    {
        MainWindow? mainWindow =
            Application.Current.Windows
                .OfType<MainWindow>()
                .FirstOrDefault();

        if (mainWindow != null)
        {
            if (_paymentCompleted)
            {
                mainWindow.ShowLockedScreen();
            }
            else
            {
                mainWindow.Show();
                mainWindow.WindowState =
                    WindowState.Normal;
                mainWindow.Activate();
                mainWindow.Focus();
            }
        }

        base.OnClosed(e);
    }
}