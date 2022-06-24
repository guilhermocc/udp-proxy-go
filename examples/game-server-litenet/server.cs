using LiteNetLib;
using LiteNetLib.Utils;

class Server
{
    static void Main(string[] args)
    {
        EventBasedNetListener listener = new EventBasedNetListener();
        NetManager server = new NetManager(listener);
        Console.WriteLine("Starting LiteNetLib Server");
        server.Start(7777 /* port */);

        listener.ConnectionRequestEvent += request =>
        {
            Console.WriteLine("Checking connected peers");
            if(server.ConnectedPeersCount < 10 /* max connections */)
                request.AcceptIfKey("SomeConnectionKey");
            else
                request.Reject();
        };

        listener.PeerConnectedEvent += peer =>
        {
            Console.WriteLine("We got connection: {0}", peer.EndPoint); // Show peer ip
            NetDataWriter writer = new NetDataWriter();                 // Create writer class
            writer.Put("Hello client!");                                // Put some string
            peer.Send(writer, DeliveryMethod.ReliableOrdered);          // Send with reliability
        };

        while (!Console.KeyAvailable)
        {
            server.PollEvents();
            Thread.Sleep(15);
        }
        Console.WriteLine("Stopping LiteNetLib Server");
        server.Stop();
    }
}

