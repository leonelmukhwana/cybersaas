using System.Speech.Synthesis;

namespace CyberSaaS.Terminal.Infrastructure.Voice;

public sealed class TerminalVoiceService : IDisposable
{
    private readonly SpeechSynthesizer _speech;
    private bool _disposed;

    public TerminalVoiceService()
    {
        _speech = new SpeechSynthesizer();

        _speech.Volume = 100;
        _speech.Rate = 0;
    }

    public void Speak(string message)
    {
        if (_disposed ||
            string.IsNullOrWhiteSpace(message))
        {
            return;
        }

        try
        {
            _speech.SpeakAsyncCancelAll();
            _speech.SpeakAsync(message);
        }
        catch
        {
            // Voice must never interrupt terminal operation.
        }
    }

    public void Dispose()
    {
        if (_disposed)
        {
            return;
        }

        _disposed = true;

        try
        {
            _speech.SpeakAsyncCancelAll();
            _speech.Dispose();
        }
        catch
        {
        }
    }
}